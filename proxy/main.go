package main

import (
	"context"
	"flag"
	"log"
	"sync"

	"github.com/ovandriyanov/dlvim/proxy/common"
	"github.com/ovandriyanov/dlvim/proxy/rpc/vim"
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

	wg.Add(1)
	go func() {
		defer wg.Done()
		vimServer := vim.NewServer()
		defer vimServer.Stop()

		vimServer.HandleClient(ctx, common.NewStdioConn())
		cancel()
	}()
	common.SetSignalHandler(ctx, cancel, &wg)

	wg.Wait()
	log.Println("Exit")
}
