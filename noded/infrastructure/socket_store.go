package infrastructure

import (
	"fmt"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
	"github.com/Pluslab/cyphonic/noded/usecase/repository"
)

var _ repository.SocketStore = (*socketStore)(nil)

type socketStore struct {
	sync.RWMutex
	values map[entity.InterfaceName]entity.Socket
}

func NewSocketStore(port string) (repository.SocketStore, error) {
	values := make(map[entity.InterfaceName]entity.Socket)

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("error getting interfaces: %v", err)
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, _ := iface.Addrs()

		ipv4Addrs := make([]netip.Addr, 0)
		ipv6Addrs := make([]netip.Addr, 0)

		for _, addr := range addrs {
			prefix, err := netip.ParsePrefix(addr.String())
			if err != nil {
				return nil, fmt.Errorf("error parsing prefix for interface %s: %v", iface.Name, err)
			}
			if prefix.Addr().Is4() {
				ipv4Addrs = append(ipv4Addrs, prefix.Addr())
			}
			if prefix.Addr().Is6() {
				ipv6Addrs = append(ipv6Addrs, prefix.Addr())
			}
		}

		if len(ipv4Addrs) == 0 && len(ipv6Addrs) == 0 {
			config.LogDebug("Interface %s has no valid IP addresses, skipping", iface.Name)
			continue
		}

		// FIXME: デバッグ用
		// lo < eth0 < eth1 の優先順位でソケットを選択するようにしているが、実際には使用できる I/F から自動的に判断する必要がある
		priority := 0
		switch iface.Name {
		case "lo":
			priority = -10
		case "eth0":
			priority = 10
		case "eth1":
			priority = 20
		}

		var ipv4Addr, ipv6Addr netip.Addr
		// use the IP address in the first array
		if len(ipv4Addrs) != 0 {
			ipv4Addr = ipv4Addrs[0]
		}
		for _, addr := range ipv6Addrs {
			if !addr.IsLinkLocalUnicast() {
				ipv6Addr = addr
				break
			}
		}

		s, err := entity.NewSocket(iface.Name, ipv4Addr, ipv6Addr, port, priority)
		if err != nil {
			config.LogErr("failed to create socket for interface",
				"interface", iface.Name,
				"error", err,
			)
			continue
		}

		config.LogDebug("created socket",
			"interface", iface.Name,
			"ipv4", ipv4Addr,
			"ipv6", ipv6Addr,
			"ipVersion", s.IPVersionString(),
			"priority", priority)

		values[entity.InterfaceName(iface.Name)] = s
	}

	return &socketStore{
		values: values,
	}, nil
}

func (s *socketStore) Values() map[entity.InterfaceName]entity.Socket {
	s.RLock()
	defer s.RUnlock()

	return s.values
}

// HighestPrioritySocket returns the socket with the highest priority
func (s *socketStore) HighestPrioritySocket() (entity.Socket, error) {
	s.RLock()
	defer s.RUnlock()

	var highestPrioritySocket entity.Socket
	highestPriority := -1

	for _, soc := range s.values {
		if soc.Priority() > highestPriority {
			highestPriority = soc.Priority()
			highestPrioritySocket = soc
		}
	}

	if highestPrioritySocket == nil {
		return nil, fmt.Errorf("no sockets available")
	}

	return highestPrioritySocket, nil
}

// ExpireDeadlines is a function that sets the deadline for all sockets to the current time
func (s *socketStore) ExpireDeadlines() error {
	s.Lock()
	defer s.Unlock()

	for _, soc := range s.values {
		if err := soc.SetReadDeadline(time.Now()); err != nil {
			return fmt.Errorf("error expiring deadline for socket %s: %v", soc.InterfaceName(), err)
		}
	}

	return nil
}

// ClearDeadlines is a function that resets the deadlines for all sockets
func (s *socketStore) ClearDeadlines() error {
	s.Lock()
	defer s.Unlock()

	for _, soc := range s.values {
		if err := soc.SetReadDeadline(time.Time{}); err != nil {
			return fmt.Errorf("error clearing deadline for socket %s: %v", soc.InterfaceName(), err)
		}
	}

	return nil
}

func (s *socketStore) SwitchActiveConn(localIPVersion entity.TypeLocalIPVersion) {
	s.Lock()
	defer s.Unlock()

	for _, soc := range s.values {
		soc.SwitchActiveConn(localIPVersion)
	}
}
