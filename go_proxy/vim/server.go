package vim

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/rpc"
)

func HandleClient(rootCtx context.Context, clientConn io.ReadWriteCloser) {
	defer clientConn.Close()

	ipcDone := make(chan struct{})

	srv := rpc.NewServer()
	srv.RegisterName(ServiceName, NewRPCHandler())
	go func() {
		decoder := json.NewDecoder(clientConn)
		for {
			var request [2]interface{}
			if err := decoder.Decode(&request); err != nil {
				if err == context.Canceled {
					log.Printf("Vim: canceled\n")
				} else {
					log.Printf("ERROR: cannot decode Vim request: %v\n", err)
				}
				ipcDone <- struct{}{}
				return
			}
			log.Printf("Vim request: %v\n", request)
		}
	}()

	select {
	case <-ipcDone:
		return
	case <-rootCtx.Done():
		_ = clientConn.Close()
		<-ipcDone
		return
	}
}
