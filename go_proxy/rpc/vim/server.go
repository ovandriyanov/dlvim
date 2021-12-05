package vim

import (
	"context"
	"io"
	"net/rpc"
)

func HandleClient(ctx context.Context, clientConn io.ReadWriteCloser) {
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
	case <-ctx.Done():
		_ = clientConn.Close()
		<-rpcDone
		return
	}
}
