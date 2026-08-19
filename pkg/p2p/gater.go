package p2p

import (
	"sync"

	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

type ConnectionGater struct {
	mu           sync.RWMutex
	blockedPeers map[peer.ID]bool
}

func NewConnectionGater() *ConnectionGater {
	return &ConnectionGater{
		blockedPeers: make(map[peer.ID]bool),
	}
}

func (g *ConnectionGater) BlockPeer(p peer.ID) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.blockedPeers[p] = true
}

func (g *ConnectionGater) UnblockPeer(p peer.ID) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.blockedPeers, p)
}

func (g *ConnectionGater) InterceptPeerDial(p peer.ID) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return !g.blockedPeers[p]
}

func (g *ConnectionGater) InterceptAddrDial(p peer.ID, m ma.Multiaddr) bool {
	return g.InterceptPeerDial(p)
}

func (g *ConnectionGater) InterceptAccept(n network.ConnMultiaddrs) bool {
	return true
}

func (g *ConnectionGater) InterceptSecured(dir network.Direction, p peer.ID, n network.ConnMultiaddrs) bool {
	return g.InterceptPeerDial(p)
}

func (g *ConnectionGater) InterceptUpgraded(n network.Conn) (bool, control.DisconnectReason) {
	return true, 0
}

var _ connmgr.ConnectionGater = (*ConnectionGater)(nil)
