package main

import (
	"context"
	"net"
)

func handleVimClient(rootCtx context.Context, clientConn net.Conn) {
	defer clientConn.Close()
}
