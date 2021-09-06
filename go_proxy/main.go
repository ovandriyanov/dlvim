package main

import (
	"context"
	"flag"
	"log"
	"sync"
)

const (
	dlvProxyAddr  = "localhost:8080"
	dlvListenAddr = "localhost:8888"
	vimServerAddr = "localhost:7778"
)

var initialized = make(chan struct{})

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	defer func() {
		// In case of panic
		cancel()
		wg.Wait()
	}()

	startupEventCh := make(chan struct{})

	setSignalHandler(ctx, cancel, &wg)
	startDlv(ctx, cancel, &wg, startupEventCh)
	<-startupEventCh
	setupServer(ctx, &wg, "Proxy", dlvProxyAddr, handleProxyClient)
	setupServer(ctx, &wg, "Vim", vimServerAddr, handleVimClient)
	close(initialized)

	wg.Wait()
	log.Printf("Exit\n")
}
