package vim

import (
	"context"
	"net"

	"github.com/ovandriyanov/dlvim/go_proxy/rpc/proxy"
	"github.com/ovandriyanov/dlvim/go_proxy/upstream"
	"github.com/ovandriyanov/dlvim/go_proxy/vimevent"
	"golang.org/x/xerrors"
)

type inventory struct {
	upstream    *upstream.Upstream
	proxyServer *proxy.Server
}

func (i *inventory) Stop() {
	i.proxyServer.Stop()
	i.upstream.Stop()
}

func (i *inventory) ProxyListenAddress() net.Addr {
	return i.proxyServer.ListenAddress()
}

func NewInventory(ctx context.Context, upstreamCommand upstream.Command, events chan<- vimevent.Event) (*inventory, error) {
	upstreamServer, err := upstream.New(ctx, upstreamCommand)
	if err != nil {
		return nil, xerrors.Errorf("cannot create upstream: %w", err)
	}

	proxyServer, err := proxy.NewServer(upstream.ListenAddress, events)
	if err != nil {
		upstreamServer.Stop()
		return nil, xerrors.Errorf("cannot create proxy server: %w", err)
	}

	return &inventory{
		upstream:    upstreamServer,
		proxyServer: proxyServer,
	}, nil
}
