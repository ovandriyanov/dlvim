package main

import (
	"context"
	"io"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

func handleProxyClient(rootCtx context.Context, clientConn io.ReadWriteCloser) {
	defer clientConn.Close()

	dlvConn, err := net.Dial("tcp", dlvListenAddr)
	noError(err)
	defer dlvConn.Close()
	log.Printf("Connected to DLV at %s\n", dlvListenAddr)

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
