package vim

import (
	"context"
	"io"
	"net/rpc"
)

func HandleClient(rootCtx context.Context, clientConn io.ReadWriteCloser) {
	defer clientConn.Close()

	rpcDone := make(chan struct{})
	srv := rpc.NewServer()
	srv.RegisterName(ServiceName, NewRPCHandler())
	go func() {
		srv.ServeCodec(NewRPCCodec(clientConn))
		rpcDone <- struct{}{}
	}()

	select {
	case <-rpcDone:
		return
	case <-rootCtx.Done():
		_ = clientConn.Close()
		<-rpcDone
		return
	}
}
