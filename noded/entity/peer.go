package entity

import (
	"net/netip"
	"sync"
)

type peer struct {
	sync.RWMutex
	fqdn          FQDN
	nodeType      NodeType // Peer's NodeType
	virtualIPv4   netip.Addr
	virtualIPv6   netip.Addr
	transactionID uint32
	tunnels       *Tunnels
}

type Peer interface {
	SetTunnel(pathID PathID, tunnel Tunnel)
	SetTunnels(tunnels *Tunnels)

	FQDN() FQDN
	NodeType() NodeType
	VirtualIPv4() netip.Addr
	VirtualIPv6() netip.Addr
	TransactionID() uint32
	Tunnel() (Tunnel, bool)
	Tunnels() *Tunnels
}

var _ Peer = (*peer)(nil)

type FQDN string

type NodeType uint32

const (
	NodeTypeInitiator NodeType = iota
	NodeTypeResponder
	NodeTypeLocalInitiator
	NodeTypeLocalResponder
)

func GeneratePeer(nodeType NodeType, rd *RouteDirection) (Peer, error) {
	var (
		fqdn                     FQDN
		virtualIPv4, virtualIPv6 netip.Addr
	)

	// TODO: 環境変数からNodePortを取得するようにする
	switch nodeType {
	case NodeTypeInitiator, NodeTypeLocalInitiator:
		fqdn = FQDN(rd.InitiatorFQDN)
		virtualIPv4 = netip.AddrFrom4(rd.InitiatorVirtualIPv4)
		virtualIPv6 = netip.AddrFrom16(rd.InitiatorVirtualIPv6)

	case NodeTypeResponder, NodeTypeLocalResponder:
		fqdn = FQDN(rd.ResponderFQDN)
		virtualIPv4 = netip.AddrFrom4(rd.ResponderVirtualIPv4)
		virtualIPv6 = netip.AddrFrom16(rd.ResponderVirtualIPv6)
	}

	return &peer{
		fqdn:          fqdn,
		nodeType:      nodeType,
		virtualIPv4:   virtualIPv4,
		virtualIPv6:   virtualIPv6,
		transactionID: rd.BaseHeader.TransactionID,
		tunnels:       NewTunnels(),
	}, nil
}

func (p *peer) SetTunnel(pathID PathID, tunnel Tunnel) {
	tunnels := p.Tunnels()
	tunnels.Set(pathID, tunnel)
}

func (p *peer) SetTunnels(tunnels *Tunnels) {
	p.Lock()
	defer p.Unlock()

	p.tunnels = tunnels
}

func (p *peer) Tunnel() (Tunnel, bool) {
	tunnels := p.Tunnels()

	return tunnels.GetRandom()
}

func (p *peer) Tunnels() *Tunnels {
	p.RLock()
	defer p.RUnlock()

	return p.tunnels
}

func (p *peer) FQDN() FQDN {
	p.RLock()
	defer p.RUnlock()

	return p.fqdn
}

func (p *peer) NodeType() NodeType {
	p.RLock()
	defer p.RUnlock()

	return p.nodeType
}

func (p *peer) VirtualIPv4() netip.Addr {
	p.RLock()
	defer p.RUnlock()

	return p.virtualIPv4
}

func (p *peer) VirtualIPv6() netip.Addr {
	p.RLock()
	defer p.RUnlock()

	return p.virtualIPv6
}

func (p *peer) TransactionID() uint32 {
	p.RLock()
	defer p.RUnlock()

	return p.transactionID
}
