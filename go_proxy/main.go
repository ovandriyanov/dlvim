package main

import (
	"context"
	"flag"
	"io"
	"log"
	"sync"

	"github.com/ovandriyanov/dlvim/go_proxy/common"
	"github.com/ovandriyanov/dlvim/go_proxy/rpc"
	"github.com/ovandriyanov/dlvim/go_proxy/rpc/dlv"
	"github.com/ovandriyanov/dlvim/go_proxy/rpc/vim"
	"github.com/ovandriyanov/dlvim/go_proxy/upstream"
)

const (
	dlvProxyAddr  = "localhost:8080"
	dlvListenAddr = "localhost:8888"
	vimServerAddr = "localhost:7778"
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

	upstream.StartDlv(ctx, cancel, &wg, dlvListenAddr)
	rpc.SetupServer(ctx, &wg, "DlvProxy", dlvProxyAddr, func(rootCtx context.Context, clientConn io.ReadWriteCloser) {
		dlv.HandleClient(rootCtx, clientConn, dlvListenAddr)
	})
	rpc.SetupServer(ctx, &wg, "Vim", vimServerAddr, vim.HandleClient)

	wg.Add(1)
	go func() {
		defer wg.Done()
		vim.HandleClient(ctx, common.NewStdioConn())
		cancel()
	}()
	common.SetSignalHandler(ctx, cancel, &wg)

	wg.Wait()
	log.Println("Exit")
}
