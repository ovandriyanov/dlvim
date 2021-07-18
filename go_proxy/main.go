package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

const (
	dlvProxyAddr = "localhost:8080"
)

func noError(err error) {
	if err != nil {
		panic(err)
	}
}

func callIf(f func() error, condition *bool) {
	if *condition {
		_ = f()
	}
}

func readPipe(pipeName string, pipe io.Reader, startupEventCh chan struct{}) {
	buf := make([]byte, 4096)
	var sentStartupEvent bool
	for {
		nRead, err := pipe.Read(buf)
		if err != nil {
			fmt.Printf("Cannot read %s from DLV server: %v\n", pipeName, err)
			return
		}
		strbuf := string(buf[:nRead])
		fmt.Printf("DLV server %s: %s\n", pipeName, strings.ReplaceAll(strbuf, "\n", "\\n"))

		if startupEventCh == nil || sentStartupEvent {
			continue
		}
		if strings.Contains(strbuf, "API server listening at:") {
			startupEventCh <- struct{}{}
			sentStartupEvent = true
			fmt.Printf("Startup event sent\n")
		}
	}
}

func startDlvServer(ctx context.Context, wg *sync.WaitGroup, startupEventCh chan struct{}) {
	cmd := exec.Command(
		"/home/ovandriyanov/bin/dlv",
		"exec",
		"/home/ovandriyanov/go/src/kek/main",
		"--listen",
		"127.0.0.1:8888",
		"--headless",
		"--accept-multiclient",
	)
	closePipes := true

	stdout, err := cmd.StdoutPipe()
	noError(err)
	defer callIf(stdout.Close, &closePipes)

	stderr, err := cmd.StderrPipe()
	noError(err)
	defer callIf(stderr.Close, &closePipes)

	err = cmd.Start()
	noError(err)
	fmt.Printf("dlv server started\n")

	wg.Add(1)
	go func() {
		defer wg.Done()

		var pipeWg sync.WaitGroup
		pipeWg.Add(2)
		go func() {
			defer pipeWg.Done()
			readPipe("stdout", stdout, startupEventCh)
		}()
		go func() {
			defer pipeWg.Done()
			readPipe("stderr", stderr, nil)
		}()

		<-ctx.Done()

		err := cmd.Process.Signal(syscall.SIGINT)
		noError(err)
		fmt.Printf("SIGINT sent to DLV (pid %d)\n", cmd.Process.Pid)

		err = cmd.Wait()
		noError(err)
		fmt.Printf("DLV exited: %v\n", cmd.ProcessState)
	}()
	closePipes = false
}

func setSignalHandler(ctx context.Context, cancel func(), wg *sync.WaitGroup) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	fmt.Printf("Signal handler has been set")

	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			return
		case signal := <-sigCh:
			fmt.Printf("Received signal %s, exiting\n", signal)
			cancel()
		}
	}()
}

func acceptDlvClients(ctx context.Context, listener net.Listener) {
	defer listener.Close()

	var wg sync.WaitGroup
	connectionsCh := make(chan net.Conn)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := listener.Accept()
			if ctx.Err() == nil {
				noError(err)
			}
			select {
			case connectionsCh <- conn:
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			listener.Close()
			wg.Wait()
			return
		case conn := <-connectionsCh:
			fmt.Printf("New DLV client connected\n")
			wg.Add(1)
			go func() {
				defer wg.Done()
				handleDlvClient(ctx, conn)
			}()

		}
	}
}

func relay(ctx context.Context, src, dst net.Conn, srcName, dstName string) {
	buf := make([]byte, 4096)
	for {
		nRead, err := src.Read(buf)
		if ctx.Err() != nil {
			break
		}
		if err != nil {
			if err == io.EOF {
				fmt.Printf("%s disconnected\n", srcName)
			} else {
				fmt.Printf("Cannot read data from %s: %v\n", srcName, err)
			}
			break
		}
		noError(err)
		strbuf := string(buf[:nRead])
		logData := strings.ReplaceAll(strbuf, "\n", "\\n")
		fmt.Printf("%s -> PRX: %s\n", srcName, logData)

		_, err = dst.Write(buf[:nRead])
		noError(err)
		fmt.Printf("PRX -> %s: %s\n", dstName, logData)
	}
	fmt.Printf("Done relaying from %s to %s\n", srcName, dstName)
}

func handleDlvClient(rootCtx context.Context, clientConn net.Conn) {
	defer clientConn.Close()
	proxyPairCtx, cancel := context.WithCancel(context.Background())

	dlvConn, err := net.Dial("tcp", "127.0.0.1:8888")
	noError(err)
	defer dlvConn.Close()
	fmt.Printf("Connected to DLV at 127.0.0.1:8888\n")

	anyoneDoneCh := make(chan struct{})
	notifyDone := func() {
		select {
		case <-proxyPairCtx.Done():
		case anyoneDoneCh <- struct{}{}:
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		relay(proxyPairCtx, dlvConn, clientConn, "DLV", "CLT")
		notifyDone()
	}()
	go func() {
		defer wg.Done()
		relay(proxyPairCtx, clientConn, dlvConn, "CLT", "DLV")
		notifyDone()
	}()

	select {
	case <-rootCtx.Done():
		cancel()
	case <-anyoneDoneCh:
		cancel()
	}
	clientConn.Close()
	dlvConn.Close()
	wg.Wait()
}

func setupProxyServer(ctx context.Context, wg *sync.WaitGroup) {
	listener, err := net.Listen("tcp", dlvProxyAddr)
	noError(err)
	fmt.Printf("Proxy server is listening at %v\n", dlvProxyAddr)

	wg.Add(1)
	go func() {
		defer wg.Done()
		acceptDlvClients(ctx, listener)
	}()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	defer func() {
		// In case of panic
		cancel()
		wg.Wait()
	}()

	startupEventCh := make(chan struct{})

	setSignalHandler(ctx, cancel, &wg)
	startDlvServer(ctx, &wg, startupEventCh)
	<-startupEventCh
	setupProxyServer(ctx, &wg)

	wg.Wait()
	fmt.Printf("Exit\n")
}
