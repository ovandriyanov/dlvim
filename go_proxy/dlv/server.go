package dlv

import (
	"context"
	"io"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/ovandriyanov/dlvim/go_proxy/common"
)

func HandleClient(rootCtx context.Context, clientConn io.ReadWriteCloser, dlvListenAddr string) {
	defer clientConn.Close()

	dlvConn, err := net.Dial("tcp", dlvListenAddr)
	common.NoError(err)
	defer dlvConn.Close()
	log.Printf("Connected to Dlv at %s\n", dlvListenAddr)

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
