package proxy

import (
	"context"
	"io"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sync"

	"github.com/ovandriyanov/dlvim/go_proxy/common"
	"github.com/ovandriyanov/dlvim/go_proxy/rpc/dlv"
	"github.com/ovandriyanov/dlvim/go_proxy/vimevent"
	"golang.org/x/xerrors"
)

type Server struct {
	dlvAddress  string
	listener    net.Listener
	acceptErrCh chan error
	ctx         context.Context
	cancelCtx   func()
	clientsWg   sync.WaitGroup
	events      chan<- vimevent.Event
}

func (s *Server) Stop() {
	if err := s.listener.Close(); err != nil {
		log.Printf("Cannot close proxy listener: %v\n", err)
	}
	common.DrainChannel(s.acceptErrCh) // No more clients are accepted

	// Wait until all client handlers are done
	s.cancelCtx()
	s.clientsWg.Wait()
}

func (s *Server) Error() <-chan error {
	return s.acceptErrCh
}

func (s *Server) ListenAddress() net.Addr {
	return s.listener.Addr()
}

func (s *Server) acceptClients() {
	defer close(s.acceptErrCh)
	for {
		clientConn, err := s.listener.Accept()
		if err != nil {
			s.acceptErrCh <- xerrors.Errorf("accept: %w", err)
			return
		}
		log.Printf("Proxy: client accepted: %s\n", clientConn.RemoteAddr().String())

		s.clientsWg.Add(1)
		go func() {
			defer s.clientsWg.Done()
			defer clientConn.Close()
			defer log.Printf("Done handling client\n")
			s.handleClient(s.ctx, clientConn)
		}()
	}
}

func (s *Server) handleClient(ctx context.Context, clientConn io.ReadWriteCloser) {
	dlvConn, err := net.Dial("tcp", s.dlvAddress)
	if err != nil {
		log.Printf("ERROR: cannot connect to dlv at %s: %v\n", s.dlvAddress, err)
		return
	}
	defer dlvConn.Close()
	log.Printf("Connected to Dlv at %s\n", s.dlvAddress)

	dlvClient := jsonrpc.NewClient(dlvConn)
	srv := rpc.NewServer()
	srv.RegisterName(dlv.ServiceName, NewRPCHandler(dlvClient, s.events, ctx))
	rpcDone := make(chan struct{})
	go func() {
		srv.ServeCodec(NewRPCCodec(clientConn, dlvClient))
		rpcDone <- struct{}{}
	}()

	select {
	case <-rpcDone:
		return
	case <-ctx.Done():
		clientConn.Close()
		<-rpcDone
	}
}

func NewServer(dlvAddress string, events chan<- vimevent.Event) (*Server, error) {
	listener, err := net.Listen("tcp", "localhost:")
	if err != nil {
		return nil, xerrors.Errorf("listen: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	server := &Server{
		dlvAddress:  dlvAddress,
		listener:    listener,
		acceptErrCh: make(chan error),
		ctx:         ctx,
		cancelCtx:   cancel,
		clientsWg:   sync.WaitGroup{},
		events:      events,
	}
	go server.acceptClients()

	return server, nil
}
