package main

import (
	"context"
	"flag"
	"io"
	"log"
	"sync"

	"github.com/ovandriyanov/dlvim/go_proxy/common"
	"github.com/ovandriyanov/dlvim/go_proxy/rpc"
	"github.com/ovandriyanov/dlvim/go_proxy/rpc/proxy"
	"github.com/ovandriyanov/dlvim/go_proxy/rpc/vim"
	"github.com/ovandriyanov/dlvim/go_proxy/upstream"
)

const (
	proxyListenAddress = "localhost:8080"
)

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	defer func() {
		// In case of panic
		cancel()
		wg.Wait()
	}()

	rpc.SetupServer(ctx, &wg, "Proxy", proxyListenAddress, func(rootCtx context.Context, clientConn io.ReadWriteCloser) {
		proxy.HandleClient(rootCtx, clientConn, upstream.ListenAddress)
	})

	wg.Add(1)
	go func() {
		defer wg.Done()
		vimServer := vim.NewServer()
		defer func() {
			if err := vimServer.Close(); err != nil {
				log.Printf("ERROR: cannot close vim server: %s", err.Error())
			}
		}()
		vimServer.HandleClient(ctx, common.NewStdioConn())
		cancel()
	}()
	common.SetSignalHandler(ctx, cancel, &wg)

	wg.Wait()
	log.Println("Exit")
}
