package upstream

import (
	"bufio"
	"context"
	"io"
	"log"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/ovandriyanov/dlvim/go_proxy/common"
	"golang.org/x/xerrors"
)

type Upstream struct {
	cmd          *exec.Cmd
	logWg        *sync.WaitGroup
	processErrCh <-chan error
}

const (
	ListenAddress = "localhost:8888"
)

func (u *Upstream) Stop() error {
	if err := u.cmd.Process.Signal(syscall.SIGINT); err != nil {
		return xerrors.Errorf("cannot send SIGINT to process %d: %w", u.cmd.Process.Pid, err)
	}
	defer common.DrainChannel(u.processErrCh)
	defer u.logWg.Wait()

	if err := <-u.processErrCh; err != nil {
		var exitErr *exec.ExitError
		if xerrors.As(err, &exitErr) {
			log.Printf("dlv terminated with non-zero status: %s", exitErr.Error())
		} else {
			log.Printf("ERROR: cannot wait dlv termination: %s", exitErr.Error())
		}
	}

	return nil
}

func (u *Upstream) Error() <-chan error {
	return u.processErrCh
}

func readPipe(pipeName string, pipe *bufio.Scanner) {
	for pipe.Scan() {
		line := pipe.Text()
		log.Printf("Upstream %s: %s\n", pipeName, line)
	}
	if pipe.Err() != nil {
		log.Printf("%s: %s", pipeName, pipe.Err())
	} else {
		log.Printf("%s: EOF", pipeName)
	}
}

func constructArgv(command Command) []string {
	var argv []string
	argv = append(argv, command.Argv()...)
	argv = append(argv, "--listen")
	argv = append(argv, ListenAddress)
	argv = append(argv, "--headless")
	argv = append(argv, "--accept-multiclient")

	return argv
}

func createPipes(cmd *exec.Cmd) (stdout, stderr io.ReadCloser, err error) {
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return nil, nil, xerrors.Errorf("stdout: %w", err)
	}

	stderr, err = cmd.StderrPipe()
	if err != nil {
		_ = stdout.Close()
		return nil, nil, xerrors.Errorf("stderr: %w", err)
	}

	return stdout, stderr, nil
}

func waitExit(cmd *exec.Cmd, exitCh chan<- error) {
	err := cmd.Wait()
	if err != nil {
		log.Printf("process %d: %s", cmd.Process.Pid, err)
	}
	exitCh <- err
}

func terminate(cmd *exec.Cmd) error {
	if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
		return xerrors.Errorf("cannot send SIGINT to process %d: %w", cmd.Process.Pid, err)
	}
	log.Printf("SIGINT sent to Dlv (pid %d)\n", cmd.Process.Pid)

	if err := cmd.Wait(); err != nil {
		return xerrors.Errorf("cannot wait process %d: %w", cmd.Process.Pid, err)
	}
	return nil
}

func readLog(stdout, stderr *bufio.Scanner) *sync.WaitGroup {
	logWg := new(sync.WaitGroup)
	logWg.Add(2)
	go func() {
		defer logWg.Done()
		readPipe("stdout", stdout)
	}()
	go func() {
		defer logWg.Done()
		readPipe("stderr", stderr)
	}()
	return logWg
}

func waitPid(cmd *exec.Cmd) <-chan error {
	processErrCh := make(chan error)
	go func() {
		defer close(processErrCh)
		if err := cmd.Wait(); err != nil {
			processErrCh <- err
		}
	}()
	return processErrCh
}

func readInitializationLine(stdoutScanner *bufio.Scanner) error {
	if !stdoutScanner.Scan() {
		if err := stdoutScanner.Err(); err != nil {
			return xerrors.Errorf("cannot scan first line from stdout: %w", err)
		}
		return xerrors.Errorf("cannot scan first line from stdout: EOF")
	}
	if !strings.Contains(stdoutScanner.Text(), "API server listening at:") {
		return xerrors.Errorf("unexpected first stdout line: %s", stdoutScanner.Text())
	}
	return nil
}

func waitDlvInitialized(ctx context.Context, cmd *exec.Cmd, stdout io.ReadCloser) (*bufio.Scanner, error) {
	stdoutScanner := bufio.NewScanner(stdout)
	dlvInitCh := make(chan error)

	defer common.DrainChannel(dlvInitCh)
	go func() {
		defer close(dlvInitCh)
		if err := readInitializationLine(stdoutScanner); err != nil {
			dlvInitCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		_ = stdout.Close()
		return nil, context.Canceled
	case err := <-dlvInitCh:
		if err != nil {
			return nil, err
		}
		return stdoutScanner, nil
	}
}

func handeInitializationFailure(initializationError error, cmd *exec.Cmd) (errorMessage string) {
	terminateErr := terminate(cmd)
	if terminateErr == nil {
		return initializationError.Error()
	}
	var exitErr *exec.ExitError
	if xerrors.As(terminateErr, &exitErr) {
		if strings.Contains(exitErr.Error(), "interrupt") {
			return initializationError.Error() // that is more descriptive
		}
		return exitErr.Error()
	}
	log.Printf("ERROR: cannot terminate dlv: %s", terminateErr.Error())
	return initializationError.Error()
}

func New(ctx context.Context, command Command) (*Upstream, error) {
	cmd := exec.Command("dlv", constructArgv(command)...)

	stdout, stderr, err := createPipes(cmd)
	if err != nil {
		return nil, xerrors.Errorf("cannot create pipes: %w", err)
	}

	if err := cmd.Start(); err != nil {
		_ = stdout.Close()
		_ = stderr.Close()
		return nil, xerrors.Errorf("cannot start process: %w", err)
	}
	log.Printf("upstream started, pid %d\n", cmd.Process.Pid)

	stdoutScanner, err := waitDlvInitialized(ctx, cmd, stdout)
	if err != nil {
		errorMessage := handeInitializationFailure(err, cmd)
		return nil, xerrors.New(errorMessage)
	}

	return &Upstream{
		cmd:          cmd,
		logWg:        readLog(stdoutScanner, bufio.NewScanner(stderr)),
		processErrCh: waitPid(cmd),
	}, nil
}
