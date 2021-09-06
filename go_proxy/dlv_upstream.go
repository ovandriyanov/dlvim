package main

import (
	"context"
	"io"
	"log"
	"os/exec"
	"strings"
	"sync"
	"syscall"
)

func readPipe(pipeName string, pipe io.Reader, startupEventCh chan struct{}) {
	buf := make([]byte, 4096)
	var sentStartupEvent bool
	for {
		nRead, err := pipe.Read(buf)
		if err != nil {
			log.Printf("Cannot read %s from DLV server: %v\n", pipeName, err)
			return
		}
		strbuf := string(buf[:nRead])
		log.Printf("DLV server %s: %s\n", pipeName, strings.ReplaceAll(strbuf, "\n", "\\n"))

		if startupEventCh == nil || sentStartupEvent {
			continue
		}
		if strings.Contains(strbuf, "API server listening at:") {
			startupEventCh <- struct{}{}
			sentStartupEvent = true
			log.Printf("Startup event sent\n")
		}
	}
}

func callIf(f func() error, condition *bool) {
	if *condition {
		_ = f()
	}
}

func startDlv(ctx context.Context, cancel func(), wg *sync.WaitGroup, startupEventCh chan struct{}) {
	cmd := exec.Command(
		"/home/ovandriyanov/go/bin/dlv",
		"exec",
		"/home/ovandriyanov/github/ovandriyanov/dlvim/helloworld/helloworld",
		"--listen",
		dlvListenAddr,
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
	log.Printf("dlv server started\n")

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
		pipesClosed := make(chan struct{})
		go func() {
			pipeWg.Wait()
			close(pipesClosed)
		}()

		select {
		case <-ctx.Done():
		case <-pipesClosed:
		}

		err := cmd.Process.Signal(syscall.SIGINT)
		noError(err)
		log.Printf("SIGINT sent to DLV (pid %d)\n", cmd.Process.Pid)

		err = cmd.Wait()
		noError(err)
		log.Printf("DLV exited: %v\n", cmd.ProcessState)
		<-pipesClosed
		cancel()
	}()
	closePipes = false
}
