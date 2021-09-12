package main

import (
	"context"
	"flag"
	"io"
	"log"
	"sync"

	"github.com/ovandriyanov/dlvim/go_proxy/common"
	"github.com/ovandriyanov/dlvim/go_proxy/dlv"
	"github.com/ovandriyanov/dlvim/go_proxy/vim"
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

	dlv.StartDlv(ctx, cancel, &wg, dlvListenAddr)
	common.SetupServer(ctx, &wg, "DlvProxy", dlvProxyAddr, func(rootCtx context.Context, clientConn io.ReadWriteCloser) {
		dlv.HandleClient(rootCtx, clientConn, dlvListenAddr)
	})
	common.SetupServer(ctx, &wg, "Vim", vimServerAddr, vim.HandleClient)

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
