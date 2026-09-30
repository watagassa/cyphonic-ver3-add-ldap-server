package repository

import (
	"net/netip"

	"github.com/Pluslab/cyphonic/noded/entity"
)

type PeerRegistry interface {
	UpsertPeer(peer entity.Peer)
	AssignTunnel(tunnel entity.Tunnel) error
	GetTunnelByPathID(pathID []byte) (entity.Tunnel, bool)
	GetPeerByVirtualAddr(vaddr netip.Addr) (entity.Peer, bool)
	GetPeerByFQDN(fqdn string) (entity.Peer, bool)
}
