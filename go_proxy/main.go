package main

import (
	"context"
	"io"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
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

func startDlvServer(ctx context.Context, cancel func(), wg *sync.WaitGroup, startupEventCh chan struct{}) {
	cmd := exec.Command(
		"/home/ovandriyanov/go/bin/dlv",
		"exec",
		"/home/ovandriyanov/github/ovandriyanov/dlvim/helloworld/helloworld",
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
			pipesClosed <- struct{}{}
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
		cancel()
	}()
	closePipes = false
}

func setSignalHandler(ctx context.Context, cancel func(), wg *sync.WaitGroup) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Signal handler has been set\n")

	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			return
		case signal := <-sigCh:
			log.Printf("Received signal %s, exiting\n", signal)
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
			log.Printf("New DLV client connected\n")
			wg.Add(1)
			go func() {
				defer wg.Done()
				handleDlvClient(ctx, conn)
			}()

		}
	}
}

func handleDlvClient(rootCtx context.Context, clientConn net.Conn) {
	defer clientConn.Close()

	dlvConn, err := net.Dial("tcp", "127.0.0.1:8888")
	noError(err)
	defer dlvConn.Close()
	log.Printf("Connected to DLV at 127.0.0.1:8888\n")

	dlvClient := jsonrpc.NewClient(dlvConn)
	srv := rpc.NewServer()
	srv.RegisterName(ServiceName, NewRPCHandler(dlvClient))
	rpcDone := make(chan struct{})
	go func() {
		srv.ServeCodec(NewRPCCodec(clientConn, dlvClient))
		rpcDone <- struct{}{}
	}()

	select {
	case <-rpcDone:
		return
	case <-rootCtx.Done():
		clientConn.Close()
		<-rpcDone
	}
}

func setupProxyServer(ctx context.Context, wg *sync.WaitGroup) {
	listener, err := net.Listen("tcp", dlvProxyAddr)
	noError(err)
	log.Printf("Proxy server is listening at %v\n", dlvProxyAddr)

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
	startDlvServer(ctx, cancel, &wg, startupEventCh)
	<-startupEventCh
	setupProxyServer(ctx, &wg)

	wg.Wait()
	log.Printf("Exit\n")
}
