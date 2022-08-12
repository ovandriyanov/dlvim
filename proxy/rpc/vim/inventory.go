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
	stack           *Stack
}

func (i *inventory) Stop() {
	i.proxyServer.Stop()
	i.upstreamClient.Close()
	if i.upstreamProcess != nil {
		i.upstreamProcess.Stop()
	}
}

func (i *inventory) ProxyListenAddress() net.Addr {
	return i.proxyServer.ListenAddress()
}

func NewInventory(ctx context.Context, upstreamStartOption upstream.StartOption, events chan<- vimevent.Event, debugRPC bool) (*inventory, error) {
	var upstreamProcess *upstream.Upstream
	var err error
	var listenAddress string
	switch startOption := upstreamStartOption.(type) {
	case upstream.StartDlvProcess:
		upstreamProcess, err = upstream.New(ctx, startOption)
		if err != nil {
			return nil, xerrors.Errorf("cannot create upstream: %w", err)
		}
		listenAddress = upstream.ListenAddress
	case *upstream.Connect:
		listenAddress = startOption.Address()
	}

	stopUpstream := func() {
		if upstreamProcess != nil {
			upstreamProcess.Stop()
		}
	}

	upstreamConnection, err := net.Dial("tcp", listenAddress)
	if err != nil {
		stopUpstream()
		return nil, xerrors.Errorf("cannot dial upstream at %s: %w", listenAddress, err)
	}

	proxyServer, err := proxy.NewServer(listenAddress, events, debugRPC)
	if err != nil {
		_ = upstreamConnection.Close()
		stopUpstream()
		return nil, xerrors.Errorf("cannot create proxy server: %w", err)
	}

	var upstreamClient commonrpc.Client = jsonrpc.NewClient(upstreamConnection)
	if debugRPC {
		upstreamClient = commonrpc.NewLoggingClient("dlv", upstreamClient)
	}

	stack := NewStack()

	return &inventory{
		upstreamProcess: upstreamProcess,
		upstreamClient:  upstreamClient,
		proxyServer:     proxyServer,
		stack:           stack,
	}, nil
}
