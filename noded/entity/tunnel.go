package entity

import (
	"net"
	"net/netip"
	"sync"
)

type Tunnel interface {
	UpdateEndpointAddr(peerAddr *net.UDPAddr) error
	RouteOptimized()

	FQDN() FQDN
	PathID() []byte
	Socket() Socket
	Endpointv4() *net.UDPAddr
	Endpointv6() *net.UDPAddr
	EndpointTRSv4() *net.UDPAddr
	EndpointTRSv6() *net.UDPAddr
	ViaTRS() bool
	EndKey() []byte
}

type tunnel struct {
	sync.RWMutex
	fqdn          FQDN
	pathID        []byte
	socket        Socket
	endpointv4    net.UDPAddr
	endpointv6    net.UDPAddr
	endpointTRSv4 net.UDPAddr
	endpointTRSv6 net.UDPAddr
	viaTRS        bool
	endKey        []byte
}

type PathID string

var _ Tunnel = (*tunnel)(nil)

func GenerateTunnel(nodeType NodeType, rd *RouteDirection, socket Socket, endKey []byte) Tunnel {
	var (
		endpointv4, endpointv6 net.UDPAddr
		fqdn                   FQDN
		viaTRS                 bool
	)

	switch nodeType {
	case NodeTypeInitiator:
		// Initiator's NAT IP address
		fqdn = FQDN(rd.InitiatorFQDN)
		endpointv4 = net.UDPAddr{
			IP:   netip.AddrFrom4(rd.NATInitiatorIPv4).AsSlice(),
			Port: int(rd.NATInitiatorPort),
		}
		endpointv6 = net.UDPAddr{
			IP:   netip.AddrFrom16(rd.NATInitiatorIPv6).AsSlice(),
			Port: int(rd.NATInitiatorPort),
		}
	case NodeTypeResponder:
		// Responder's NAT IP address
		fqdn = FQDN(rd.ResponderFQDN)
		endpointv4 = net.UDPAddr{
			IP:   netip.AddrFrom4(rd.NATResponderIPv4).AsSlice(),
			Port: int(rd.NATResponderPort),
		}
		endpointv6 = net.UDPAddr{
			IP:   netip.AddrFrom16(rd.NATResponderIPv6).AsSlice(),
			Port: int(rd.NATResponderPort),
		}
	case NodeTypeLocalInitiator:
		// Initiator's physical IP address
		fqdn = FQDN(rd.InitiatorFQDN)
		endpointv4 = net.UDPAddr{
			IP:   netip.AddrFrom4(rd.InitiatorRealIPv4).AsSlice(),
			Port: 30000,
		}
		endpointv6 = net.UDPAddr{
			IP:   netip.AddrFrom16(rd.InitiatorRealIPv6).AsSlice(),
			Port: 30000,
		}
	case NodeTypeLocalResponder:
		// Responder's physical IP address
		fqdn = FQDN(rd.ResponderFQDN)
		endpointv4 = net.UDPAddr{
			IP:   netip.AddrFrom4(rd.ResponderRealIPv4).AsSlice(),
			Port: 30000,
		}
		endpointv6 = net.UDPAddr{
			IP:   netip.AddrFrom16(rd.ResponderRealIPv6).AsSlice(),
			Port: 30000,
		}
	}

	endpointTRSv4 := net.UDPAddr{
		IP:   rd.TRSIPv4[:],
		Port: 4507, // FIXME: 固定値
	}
	endpointTRSv6 := net.UDPAddr{
		IP:   rd.TRSIPv6[:],
		Port: 4507, // FIXME: 固定値
	}

	if rd.ProcessCode == TunnelRequestToTRS {
		viaTRS = true
	} else {
		viaTRS = false
	}

	pathID := make([]byte, 16)
	copy(pathID, rd.PathID)
	end := make([]byte, 16)
	copy(end, endKey)

	return &tunnel{
		fqdn:          fqdn,
		pathID:        pathID,
		socket:        socket,
		endpointv4:    endpointv4,
		endpointv6:    endpointv6,
		endpointTRSv4: endpointTRSv4,
		endpointTRSv6: endpointTRSv6,
		viaTRS:        viaTRS,
		endKey:        end,
	}
}

func (t *tunnel) UpdateEndpointAddr(peerAddr *net.UDPAddr) error {
	t.Lock()
	defer t.Unlock()

	addr := peerAddr.AddrPort().Addr().Unmap()

	switch {
	case addr.Is4():
		t.endpointv4 = *peerAddr
	case addr.Is6():
		t.endpointv6 = *peerAddr
	default:
		return net.InvalidAddrError("invalid IP address")
	}
	return nil
}

func (t *tunnel) RouteOptimized() {
	t.Lock()
	defer t.Unlock()

	// TODO: AtomicBoolに変更する
	t.viaTRS = false
}

func (t *tunnel) FQDN() FQDN {
	t.RLock()
	defer t.RUnlock()

	return t.fqdn
}

func (t *tunnel) PathID() []byte {
	t.RLock()
	defer t.RUnlock()

	return t.pathID
}

func (t *tunnel) Socket() Socket {
	t.RLock()
	defer t.RUnlock()

	return t.socket
}

func (t *tunnel) Endpointv4() *net.UDPAddr {
	t.RLock()
	defer t.RUnlock()

	return &t.endpointv4
}

func (t *tunnel) Endpointv6() *net.UDPAddr {
	t.RLock()
	defer t.RUnlock()

	return &t.endpointv6
}

func (t *tunnel) EndpointTRSv4() *net.UDPAddr {
	t.RLock()
	defer t.RUnlock()

	return &t.endpointTRSv4
}

func (t *tunnel) EndpointTRSv6() *net.UDPAddr {
	t.RLock()
	defer t.RUnlock()

	return &t.endpointTRSv6
}

func (t *tunnel) ViaTRS() bool {
	t.RLock()
	defer t.RUnlock()

	return t.viaTRS
}

func (t *tunnel) EndKey() []byte {
	t.RLock()
	defer t.RUnlock()

	return t.endKey
}
