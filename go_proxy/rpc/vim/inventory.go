package vim

import (
	"context"

	"github.com/ovandriyanov/dlvim/go_proxy/rpc/proxy"
	"github.com/ovandriyanov/dlvim/go_proxy/upstream"
	"golang.org/x/xerrors"
)

type inventory struct {
	upstream    *upstream.Upstream
	proxyServer *proxy.Server

	proxyListenAddress string
}

func (i *inventory) Stop() {
	i.proxyServer.Stop()
	i.upstream.Stop()
}

func NewInventory(ctx context.Context, upstreamCommand upstream.Command) (*inventory, error) {
	upstreamServer, err := upstream.New(ctx, upstreamCommand)
	if err != nil {
		return nil, xerrors.Errorf("cannot create upstream: %w", err)
	}

	proxyServer, proxyListenAddress, err := proxy.NewServer(upstream.ListenAddress)
	if err != nil {
		upstreamServer.Stop()
		return nil, xerrors.Errorf("cannot create proxy server: %w", err)
	}

	return &inventory{
		upstream:           upstreamServer,
		proxyServer:        proxyServer,
		proxyListenAddress: proxyListenAddress.String(),
	}, nil
}
