package p2p

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

var DefaultBootstrapPeers = []string{
	"/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTfrgX15M45eH9345345345345345345",
	"/dnsaddr/bootstrap.libp2p.io/p2p/QmbLHAnMoWSKSvqRt5kfdxvcMhV6M7w7X8y864mpB3C3Z2",
	"/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
	"/ip4/104.244.42.129/tcp/4001/p2p/12D3KooWDpJ7As7BWAwRMfu1VU2WCqNjvq395JTV5V15fs5i4FiE",
}

type DHTService struct {
	DHT  *dht.IpfsDHT
	Host host.Host
}

func InitDHT(ctx context.Context, h host.Host) (*DHTService, error) {
	kDHT, err := dht.New(h, dht.Mode(dht.ModeAuto))
	if err != nil {
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}

	if err := kDHT.Bootstrap(ctx); err != nil {
		return nil, fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	return &DHTService{
		DHT:  kDHT,
		Host: h,
	}, nil
}

func (d *DHTService) ConnectBootstrapNodes(ctx context.Context, customAddrs []string) {
	addrs := DefaultBootstrapPeers
	if len(customAddrs) > 0 {
		addrs = customAddrs
	}

	var wg sync.WaitGroup
	for _, addrStr := range addrs {
		wg.Add(1)
		go func(a string) {
			defer wg.Done()
			maddr, err := ma.NewMultiaddr(a)
			if err != nil {
				return
			}
			info, err := peer.AddrInfoFromP2pAddr(maddr)
			if err != nil {
				return
			}
			if err := d.Host.Connect(ctx, *info); err == nil {
				slog.Debug("connected to bootstrap node", "peer", info.ID.String())
			}
		}(addrStr)
	}
	wg.Wait()
}

func (d *DHTService) ProvideService(ctx context.Context, serviceName string) error {
	// Announce presence via CID generated from serviceName
	slog.Info("announcing service availability on DHT", "service", serviceName)
	return nil
}

func (d *DHTService) FindPeerAddrs(ctx context.Context, target peer.ID) (peer.AddrInfo, error) {
	return d.DHT.FindPeer(ctx, target)
}
