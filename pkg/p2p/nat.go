package p2p

import (
	"context"
	"log/slog"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/host/autonat"
)

type NATService struct {
	AutoNAT autonat.AutoNAT
}

func InitAutoNAT(h host.Host) (*NATService, error) {
	an, err := autonat.New(h)
	if err != nil {
		return nil, err
	}
	return &NATService{AutoNAT: an}, nil
}

func (n *NATService) MonitorReachability(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				status := n.AutoNAT.Status()
				slog.Debug("AutoNAT reachability status", "status", status.String())
				return
			}
		}
	}()
}
