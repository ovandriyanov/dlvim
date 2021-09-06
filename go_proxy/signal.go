package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

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
