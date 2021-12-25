package main

import (
	"context"
	"flag"
	"log"
	"sync"

	"github.com/ovandriyanov/dlvim/proxy/common"
	"github.com/ovandriyanov/dlvim/proxy/rpc/vim"
)

var (
	debugRPC = flag.Bool("debug-rpc", false, "Show full requests and responses sent between dlv and the proxy")
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
		vimServer := vim.NewServer(*debugRPC)
		defer vimServer.Stop()

		vimServer.HandleClient(ctx, common.NewStdioConn())
		cancel()
	}()
	common.SetSignalHandler(ctx, cancel, &wg)

	wg.Wait()
	log.Println("Exit")
}
