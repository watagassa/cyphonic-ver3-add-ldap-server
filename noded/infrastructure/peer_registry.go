package infrastructure

import (
	"encoding/hex"
	"fmt"
	"net/netip"

	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/usecase/repository"
)

var _ repository.PeerRegistry = (*peerRegistry)(nil)

type peerRegistry struct {
	peers   *entity.Peers
	tunnels *entity.Tunnels
}

func NewPeerRegistry() repository.PeerRegistry {
	return &peerRegistry{
		peers:   entity.NewPeers(),
		tunnels: entity.NewTunnels(),
	}
}

func (r *peerRegistry) UpsertPeer(peer entity.Peer) {
	// update
	if p, ok := r.peers.Get(peer.FQDN()); ok {
		peer.SetTunnels(p.Tunnels())
	}
	// insert
	r.peers.Set(peer.FQDN(), peer)
}

func (r *peerRegistry) AssignTunnel(tunnel entity.Tunnel) error {
	peer, ok := r.peers.Get(tunnel.FQDN())
	if !ok {
		return fmt.Errorf("peer not found for FQDN: %s", tunnel.FQDN())
	}

	pathID := entity.PathID(hex.EncodeToString(tunnel.PathID()))
	peer.SetTunnel(pathID, tunnel)
	r.tunnels.Set(pathID, tunnel)

	return nil
}

func (r *peerRegistry) GetTunnelByPathID(pathID []byte) (entity.Tunnel, bool) {
	t, ok := r.tunnels.Get(entity.PathID(hex.EncodeToString(pathID)))
	return t, ok
}

func (r *peerRegistry) GetPeerByVirtualAddr(vaddr netip.Addr) (entity.Peer, bool) {
	return r.peers.GetByVirtualAddr(vaddr)
}

func (r *peerRegistry) GetPeerByFQDN(fqdn string) (entity.Peer, bool) {
	return r.peers.Get(entity.FQDN(fqdn))
}
