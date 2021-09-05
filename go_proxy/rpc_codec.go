package main

import (
	"io"
	"log"
	"net/rpc"
	"net/rpc/jsonrpc"

	"golang.org/x/xerrors"
)

type loggingServerCodec struct {
	rpc.ServerCodec

	dlvClient *rpc.Client
}

func (c *loggingServerCodec) Close() (resultErr error) {
	if err := c.dlvClient.Close(); err != nil {
		resultErr = xerrors.Errorf("cannot close dlv client: %v", err)
	}
	if err := c.ServerCodec.Close(); err != nil {
		resultErr = xerrors.Errorf("cannot underlying server codec: %v", err)
	}
	return
}

func (c *loggingServerCodec) ReadRequestHeader(request *rpc.Request) error {
	err := c.ServerCodec.ReadRequestHeader(request)
	if err != nil {
		log.Printf("Cannot read request header: %v\n", err)
		return err
	}
	_, isKnown := KnownMethods[request.ServiceMethod]
	if !isKnown {
		log.Printf("WARNING: unknown method %s\n", request.ServiceMethod)
	}
	return nil
}

func NewRPCCodec(conn io.ReadWriteCloser, dlvClient *rpc.Client) rpc.ServerCodec {
	return &loggingServerCodec{
		ServerCodec: jsonrpc.NewServerCodec(conn),
		dlvClient:   dlvClient,
	}
}
