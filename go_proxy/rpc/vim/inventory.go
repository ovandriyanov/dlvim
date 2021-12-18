package vim

import (
	"context"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/ovandriyanov/dlvim/go_proxy/rpc/proxy"
	"github.com/ovandriyanov/dlvim/go_proxy/upstream"
	"github.com/ovandriyanov/dlvim/go_proxy/vimevent"
	"golang.org/x/xerrors"
)

type inventory struct {
	upstreamProcess *upstream.Upstream
	upstreamClient  *rpc.Client
	proxyServer     *proxy.Server
}

func (i *inventory) Stop() {
	i.proxyServer.Stop()
	i.upstreamClient.Close()
	i.upstreamProcess.Stop()
}

func (i *inventory) ProxyListenAddress() net.Addr {
	return i.proxyServer.ListenAddress()
}

func NewInventory(ctx context.Context, upstreamCommand upstream.Command, events chan<- vimevent.Event) (*inventory, error) {
	upstreamProcess, err := upstream.New(ctx, upstreamCommand)
	if err != nil {
		return nil, xerrors.Errorf("cannot create upstream: %w", err)
	}

	upstreamConnection, err := net.Dial("tcp", upstream.ListenAddress)
	if err != nil {
		upstreamProcess.Stop()
		return nil, xerrors.Errorf("cannot dial upstream at %s: %w", upstream.ListenAddress, err)
	}

	proxyServer, err := proxy.NewServer(upstream.ListenAddress, events)
	if err != nil {
		_ = upstreamConnection.Close()
		upstreamProcess.Stop()
		return nil, xerrors.Errorf("cannot create proxy server: %w", err)
	}

	return &inventory{
		upstreamProcess: upstreamProcess,
		upstreamClient:  jsonrpc.NewClient(upstreamConnection),
		proxyServer:     proxyServer,
	}, nil
}
