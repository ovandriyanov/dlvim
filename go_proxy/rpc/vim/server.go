package vim

import (
	"context"
	"io"
	"net/rpc"
	"sync"
	"time"

	"github.com/ovandriyanov/dlvim/go_proxy/upstream"
	"golang.org/x/xerrors"
)

type Server struct {
	mutex    sync.Mutex
	upstream *upstream.Upstream
}

func (s *Server) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.upstream == nil {
		return nil
	}

	return s.upstream.Stop()
}

func (s *Server) HandleClient(ctx context.Context, clientConn io.ReadWriteCloser) {
	defer clientConn.Close()

	rpcDone := make(chan struct{})
	srv := rpc.NewServer()
	rpcHandler := NewRPCHandler(s)
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

func (s *Server) InitializeUpstream(command upstream.Command) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.upstream != nil {
		return xerrors.New("already initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s.upstream, err = upstream.Start(ctx, command)
	if err != nil {
		return xerrors.Errorf("cannot start dlv: %w", err)
	}

	return nil
}

func NewServer() *Server {
	return &Server{
		mutex:    sync.Mutex{},
		upstream: nil,
	}
}
