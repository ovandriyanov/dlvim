package rpc

import (
	"context"
	"io"
	"log"
	"net"
	"sync"

	"github.com/ovandriyanov/dlvim/go_proxy/common"
)

type clientHandler func(rootCtx context.Context, clientConn io.ReadWriteCloser)

func acceptClients(ctx context.Context, listener net.Listener, name string, handler clientHandler) {
	defer listener.Close()

	var wg sync.WaitGroup
	connectionsCh := make(chan net.Conn)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := listener.Accept()
			if ctx.Err() == nil {
				common.NoError(err)
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
			log.Printf("New %s client connected\n", name)
			wg.Add(1)
			go func() {
				defer wg.Done()
				handler(ctx, conn)
			}()
		}
	}
}

func SetupServer(ctx context.Context, wg *sync.WaitGroup, name, addr string, handler clientHandler) {
	listener, err := net.Listen("tcp", addr)
	common.NoError(err)
	log.Printf("%s server is listening at %v\n", name, addr)

	wg.Add(1)
	go func() {
		defer wg.Done()
		acceptClients(ctx, listener, name, handler)
	}()
}
