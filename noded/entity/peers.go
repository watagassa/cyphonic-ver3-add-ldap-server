package entity

import (
	"net/netip"
	"sync"
)

type Peers struct {
	sync.RWMutex
	m map[FQDN]Peer
}

func NewPeers() *Peers {
	return &Peers{
		m: make(map[FQDN]Peer),
	}
}

func (ps *Peers) Get(fqdn FQDN) (Peer, bool) {
	ps.RLock()
	defer ps.RUnlock()

	peer, ok := ps.m[fqdn]
	return peer, ok
}

func (ps *Peers) Set(fqdn FQDN, peer Peer) {
	ps.Lock()
	defer ps.Unlock()

	ps.m[fqdn] = peer
}

func (ps *Peers) GetByVirtualAddr(vaddr netip.Addr) (Peer, bool) {
	ps.RLock()
	defer ps.RUnlock()

	for _, p := range ps.m {
		switch vaddr {
		case p.VirtualIPv4():
			return p, true
		case p.VirtualIPv6():
			return p, true
		}
	}
	return nil, false
}
