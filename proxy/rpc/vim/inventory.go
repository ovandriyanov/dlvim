package vim

import (
	"context"
	"net"
	"net/rpc/jsonrpc"

	commonrpc "github.com/ovandriyanov/dlvim/proxy/rpc"
	"github.com/ovandriyanov/dlvim/proxy/rpc/proxy"
	"github.com/ovandriyanov/dlvim/proxy/upstream"
	"github.com/ovandriyanov/dlvim/proxy/vimevent"
	"golang.org/x/xerrors"
)

type inventory struct {
	upstreamProcess *upstream.Upstream
	upstreamClient  commonrpc.Client
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

func NewInventory(ctx context.Context, upstreamCommand upstream.Command, events chan<- vimevent.Event, debugRPC bool) (*inventory, error) {
	upstreamProcess, err := upstream.New(ctx, upstreamCommand)
	if err != nil {
		return nil, xerrors.Errorf("cannot create upstream: %w", err)
	}

	upstreamConnection, err := net.Dial("tcp", upstream.ListenAddress)
	if err != nil {
		upstreamProcess.Stop()
		return nil, xerrors.Errorf("cannot dial upstream at %s: %w", upstream.ListenAddress, err)
	}

	proxyServer, err := proxy.NewServer(upstream.ListenAddress, events, debugRPC)
	if err != nil {
		_ = upstreamConnection.Close()
		upstreamProcess.Stop()
		return nil, xerrors.Errorf("cannot create proxy server: %w", err)
	}

	var upstreamClient commonrpc.Client = jsonrpc.NewClient(upstreamConnection)
	if debugRPC {
		upstreamClient = commonrpc.NewLoggingClient("dlv", upstreamClient)
	}

	return &inventory{
		upstreamProcess: upstreamProcess,
		upstreamClient:  upstreamClient,
		proxyServer:     proxyServer,
	}, nil
}
