package infrastructure

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/Pluslab/cyphonic/u-trsd/infrastructure/config"
	"github.com/Pluslab/cyphonic/u-trsd/usecase/repository"
)

// ErrPathinfoNotFound is an error that indicates that the path information was not found in the cache or database.
var ErrPathinfoNotFound = errors.New("path information not found")

// ErrIPAddressNotFound is an error that indicates that the IP address was not found in the cache or database.
var ErrIPAddressNotFound = errors.New("failed to get IP address")

// ErrUnknownAddressType is an error that indicates that the address type is unknown.
var ErrUnknownAddressType = errors.New("unknown address type")

// ErrFailedToUpdateNodeAddress is an error that indicates that the node address update failed.
var ErrFailedToUpdateNodeAddress = errors.New("failed to update node address")

// ErrFailedToChangeGeneratingFlag is an error that indicates that changing the generating flag failed.
var ErrFailedToChangeGeneratingFlag = errors.New("failed to change generating flag")

// pathInfo is the data to be cached.
type pathInfo struct {
	generatingFlag      bool
	inIPv4              net.IP
	inIPv6              net.IP
	inPort              uint16
	rnIPv4              net.IP
	rnIPv6              net.IP
	rnPort              uint16
	tunnelKey           []byte
	tunnelKeyCipherType uint16
	tunnelKeyLength     uint16
	expires             int64
}

// PathCache is a struct for caching.
type PathCache struct {
	pathInfos  map[string]*pathInfo
	lock       sync.RWMutex
	sqlHandler repository.SQLHandler
}

// NewPathCache initializes a path cache.
func NewPathCache(sqlHandler repository.SQLHandler) (*PathCache, error) {
	c := &PathCache{pathInfos: make(map[string]*pathInfo)}
	c.sqlHandler = sqlHandler
	go func() {
		t := time.NewTicker(10 * time.Second)
		defer t.Stop()
		for range t.C {
			c.lock.Lock()
			for k, v := range c.pathInfos {
				if v.isExpired(time.Now().UnixNano()) {
					delete(c.pathInfos, k)
				}
			}
			c.lock.Unlock()
		}
	}()
	return c, nil
}

// isExpired determines if it has expires.
func (p *pathInfo) isExpired(time int64) bool {
	if p.expires == 0 {
		return true
	}
	return time > p.expires
}

// GetResponderNodeAddr get responder node address.
func (c *PathCache) GetResponderNodeAddr(pathID string) (int, net.IP, uint16, error) {
	v, err := c.get(pathID)
	if err != nil {
		config.LogErr("PathInfo not found", "pathID", pathID)
		return 0, nil, 0, err
	}

	switch {
	case !v.rnIPv6.Equal(net.IPv6zero):
		return 6, v.rnIPv6, v.rnPort, nil
	case !v.rnIPv4.Equal(net.IPv4zero):
		return 4, v.rnIPv4, v.rnPort, nil
	}

	return 0, nil, 0, ErrIPAddressNotFound
}

// GetInitiatorNodeAddr get initiator node address.
func (c *PathCache) GetInitiatorNodeAddr(pathID string) (int, net.IP, uint16, error) {
	v, err := c.get(pathID)
	if err != nil {
		config.LogErr("PathInfo not found", "pathID", pathID)
		return 0, nil, 0, err
	}

	switch {
	case !v.inIPv6.Equal(net.IPv6zero):
		return 6, v.inIPv6, v.inPort, nil
	case !v.inIPv4.Equal(net.IPv4zero):
		return 4, v.inIPv4, v.inPort, nil
	}

	return 0, nil, 0, ErrIPAddressNotFound
}

// get gets a value from a cache.
func (c *PathCache) get(pathID string) (*pathInfo, error) {
	c.lock.RLock()
	v, exist := c.pathInfos[pathID]
	expired := exist && v.isExpired(time.Now().UnixNano())
	c.lock.RUnlock()

	if !exist || expired {
		pathInfo, err := c.getPathinfoFromDB(pathID)
		if err != nil {
			return nil, err
		}

		c.lock.Lock()
		c.pathInfos[pathID] = pathInfo
		c.lock.Unlock()

		v = pathInfo
		config.LogDebug("Path information not found in cache, fetched from DB", "pathID", pathID, "path_info", v)
	}

	return v, nil
}

func (c *PathCache) getPathinfoFromDB(pathID string) (*pathInfo, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	pathInformation, err := c.sqlHandler.GetPathInformationByPathID(pathID)
	if err != nil {
		config.LogErr("Failed to get path information from DB", "pathID", pathID, "error", err)
		return nil, ErrPathinfoNotFound
	}

	v := &pathInfo{
		generatingFlag:      pathInformation.GeneratingFlag,
		inIPv4:              pathInformation.InitiatorIPv4,
		inIPv6:              pathInformation.InitiatorIPv6,
		inPort:              pathInformation.InitiatorPort,
		rnIPv4:              pathInformation.ResponderIPv4,
		rnIPv6:              pathInformation.ResponderIPv6,
		rnPort:              pathInformation.ResponderPort,
		tunnelKey:           pathInformation.TunnelKey,
		tunnelKeyCipherType: pathInformation.TunnelKeyCipherType,
		tunnelKeyLength:     pathInformation.TunnelKeyLength,
		expires:             time.Now().Add(5 * time.Minute).UnixNano(),
	}

	if v.inIPv4 == nil {
		v.inIPv4 = net.IPv4zero
	}
	if v.inIPv6 == nil {
		v.inIPv6 = net.IPv6zero
	}
	if v.rnIPv4 == nil {
		v.rnIPv4 = net.IPv4zero
	}
	if v.rnIPv6 == nil {
		v.rnIPv6 = net.IPv6zero
	}

	if v.tunnelKey == nil {
		config.LogErr("TunnelKey is nil", "pathID", pathID)
		return nil, ErrPathinfoNotFound
	}

	// GenaratingFlag is true until receive Tunnel request.
	v.generatingFlag = true

	config.LogDebug("Fetched path information from DB", "pathID", pathID, "path_info", v)
	return v, nil
}

// Update the expiration time of the cache.
func (c *PathCache) Update(pathID string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if v, ok := c.pathInfos[pathID]; ok {
		v.expires = time.Now().Add(5 * time.Minute).UnixNano()
	}
}

// UpdateResponderNodeAddr regists the IP address and port of the Responder Node in PathInformations table.
func (c *PathCache) UpdateResponderNodeAddr(pathID string, ip net.IP, port uint16) error {
	// Convert to IPv4
	ip4 := ip.To4()

	// If IPv6, ip4 will be nil.
	if ip4 == nil {
		return c.updateResponderNodeAddr(pathID, net.IPv4zero, ip, port)
	}
	return c.updateResponderNodeAddr(pathID, ip4, net.IPv6zero, port)
}

func (c *PathCache) updateResponderNodeAddr(pathID string, ipv4, ipv6 net.IP, port uint16) error {
	v, err := c.get(pathID)
	switch {
	case err != nil:
		return err
	case !v.generatingFlag:
		config.LogDebug("Updating RN Addr but generatingFlag is false. This means path info has already updated.(Maybe TRS has already received Tunnel Request.)", "Path Informations", v)
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	config.LogDebug("Results of searching a cache to update responder node address", "path_informations", v)

	rnIPv4 := make([]byte, net.IPv4len)
	copy(rnIPv4, ipv4)
	rnIPv6 := make([]byte, net.IPv6len)
	copy(rnIPv6, ipv6)
	v.rnIPv4 = rnIPv4
	v.rnIPv6 = rnIPv6
	v.rnPort = port
	v.expires = time.Now().Add(5 * time.Minute).UnixNano()

	config.LogDebug("Updated responder node address", "pathID", pathID, "rnIPv4", v.rnIPv4, "rnIPv6", v.rnIPv6, "rnPort", v.rnPort)

	return nil
}

// UpdateInitiatorNodeAddr regists the IP address and port of the Initiator Node in PathInformations table.
func (c *PathCache) UpdateInitiatorNodeAddr(pathID string, ip net.IP, port uint16) error {
	// Convert to IPv4
	ip4 := ip.To4()

	// If IPv6, ip4 will be nil.
	if ip4 == nil {
		return c.updateInitiatorNodeAddr(pathID, net.IPv4zero, ip, port)
	}
	return c.updateInitiatorNodeAddr(pathID, ip4, net.IPv6zero, port)
}

func (c *PathCache) updateInitiatorNodeAddr(pathID string, ipv4, ipv6 net.IP, port uint16) error {
	v, err := c.get(pathID)
	switch {
	case err != nil:
		return err
	case !v.generatingFlag:
		config.LogDebug("Updating IN Addr but generatingFlag is false. This means path info has already updated.(Maybe TRS has already received Tunnel Request.)", "Path Informations", v)
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	config.LogDebug("Results of searching a cache to update initiator node address", "path_informations", v)

	inIPv4 := make([]byte, net.IPv4len)
	copy(inIPv4, ipv4)
	inIPv6 := make([]byte, net.IPv6len)
	copy(inIPv6, ipv6)
	v.inIPv4 = inIPv4
	v.inIPv6 = inIPv6
	v.inPort = port
	v.expires = time.Now().Add(5 * time.Minute).UnixNano()

	v.generatingFlag = false

	config.LogDebug("Updated initiator node address", "pathID", pathID, "inIPv4", v.inIPv4, "inIPv6", v.inIPv6, "inPort", v.inPort)

	return nil
}

// GetTunnelKey retrieves the tunnel key from the cache or database.
func (c *PathCache) GetTunnelKey(pathID string) (key []byte, keyCipherType uint16, keyLength uint16, err error) {
	pathInformation, err := c.get(pathID)
	if err != nil {
		config.LogErr("Failed to get path information", "pathID", pathID, "error", err)
		return nil, 0, 0, err
	}

	if pathInformation.tunnelKey == nil {
		config.LogErr("TunnelKey is nil", "pathID", pathID)
		return nil, 0, 0, ErrPathinfoNotFound
	}

	return pathInformation.tunnelKey, pathInformation.tunnelKeyCipherType, pathInformation.tunnelKeyLength, nil
}
