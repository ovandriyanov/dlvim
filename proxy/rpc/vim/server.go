package vim

import (
	"context"
	"io"
	"net/rpc"
	"sync"
	"time"

	commonrpc "github.com/ovandriyanov/dlvim/proxy/rpc"
	"github.com/ovandriyanov/dlvim/proxy/upstream"
	"github.com/ovandriyanov/dlvim/proxy/vimevent"
	"golang.org/x/xerrors"
)

type Server struct {
	events    chan vimevent.Event
	mutex     sync.Mutex
	inventory *inventory
	debugRPC  bool
}

func (s *Server) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.inventory == nil {
		return
	}

	s.inventory.Stop()
}

func (s *Server) UpstreamClient() commonrpc.Client {
	if s.inventory == nil {
		return nil
	}
	return s.inventory.upstreamClient
}

func (s *Server) HandleClient(ctx context.Context, clientConn io.ReadWriteCloser) {
	defer clientConn.Close()

	rpcDone := make(chan struct{})
	srv := rpc.NewServer()
	rpcHandler := NewRPCHandler(s, ctx)
	var receiver interface{} = rpcHandler
	if s.debugRPC {
		receiver = NewLoggingRPCHandler(rpcHandler, "vim server")
	}
	srv.RegisterName(ServiceName, receiver)
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

func (s *Server) Initialize(startOption upstream.StartOption) (inventory *inventory, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.inventory != nil {
		return nil, xerrors.New("already initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s.inventory, err = NewInventory(ctx, startOption, s.events, s.debugRPC)
	if err != nil {
		return nil, err
	}

	return s.inventory, nil
}

func NewServer(debugRPC bool) *Server {
	return &Server{
		events:    make(chan vimevent.Event),
		mutex:     sync.Mutex{},
		inventory: nil,
		debugRPC:  debugRPC,
	}
}
