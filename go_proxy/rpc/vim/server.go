package vim

import (
	"context"
	"io"
	"net/rpc"
	"sync"
	"time"

	"github.com/ovandriyanov/dlvim/go_proxy/upstream"
	"github.com/ovandriyanov/dlvim/go_proxy/vimevent"
	"golang.org/x/xerrors"
)

type Server struct {
	events    chan vimevent.Event
	mutex     sync.Mutex
	inventory *inventory
}

func (s *Server) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.inventory == nil {
		return
	}

	s.inventory.Stop()
}

func (s *Server) UpstreamClient() *rpc.Client {
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
	srv.RegisterName(ServiceName, rpcHandler)
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

func (s *Server) Initialize(command upstream.Command) (inventory *inventory, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.inventory != nil {
		return nil, xerrors.New("already initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s.inventory, err = NewInventory(ctx, command, s.events)
	if err != nil {
		return nil, err
	}

	return s.inventory, nil
}

func NewServer() *Server {
	return &Server{
		events:    make(chan vimevent.Event),
		mutex:     sync.Mutex{},
		inventory: nil,
	}
}
