package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"net/netip"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
	"github.com/Pluslab/cyphonic/noded/usecase/repository"
)

var _ repository.RouteDirectionCache = (*routeDirectionCache)(nil)

// routeDirectionCache is a struct for caching.
type routeDirectionCache struct {
	isReady atomic.Bool
	values  map[string]*routeDirectionInfo
	mu      sync.Mutex
}

type routeDirectionInfo struct {
	data                 []byte
	shouldUpdatePeerAddr bool
	peerAddr             netip.AddrPort
	expires              int64
}

func NewRouteDirectionCache() repository.RouteDirectionCache {
	return &routeDirectionCache{
		isReady: atomic.Bool{},
		values:  make(map[string]*routeDirectionInfo),
	}
}

func (c *routeDirectionCache) Run(ctx context.Context) error {
	config.LogDebug("Starting route direction cache")
	defer config.LogDebug("Finished route direction cache")

	t := time.NewTicker(time.Second)
	defer t.Stop()

	c.isReady.Store(true)
	defer c.isReady.Store(false)

	for {
		select {
		case <-t.C:
			c.mu.Lock()
			for k, v := range c.values {
				if v.expired(time.Now().UnixNano()) {
					delete(c.values, k)
				}
			}
			c.mu.Unlock()
		case <-ctx.Done():
			return nil
		}
	}
}

// Get gets a value from a cache.
func (c *routeDirectionCache) Get(key string) (routeDirection entity.RouteDirection, shouldUpdatePeerAddr bool, peerAddr netip.AddrPort, err error) {
	if !c.isReady.Load() {
		return entity.RouteDirection{}, false, netip.AddrPort{}, fmt.Errorf("cache is not ready")
	}

	rd := entity.RouteDirection{}
	c.mu.Lock()
	if v, ok := c.values[key]; ok {
		err := json.Unmarshal(v.data, &rd)
		if err != nil {
			return rd, false, netip.AddrPort{}, fmt.Errorf("JSON unmarshal error: %w", err)
		}
		routeDirection = rd
		shouldUpdatePeerAddr = v.shouldUpdatePeerAddr
		peerAddr = v.peerAddr
	}
	c.mu.Unlock()

	return routeDirection, shouldUpdatePeerAddr, peerAddr, nil
}

// Put puts a RouteDirection packet in a cache. If a key and value exist, overwrite it.
func (c *routeDirectionCache) Put(key string, routeDirection *entity.RouteDirection, expires int64) error {
	if !c.isReady.Load() {
		return fmt.Errorf("cache is not ready")
	}

	bytes, err := json.Marshal(&routeDirection)
	if err != nil {
		return fmt.Errorf("JSON unmarshal error: %w", err)
	}

	rd := make([]byte, len(bytes))
	copy(rd, bytes)
	c.mu.Lock()
	if v, ok := c.values[key]; ok {
		c.values[key] = &routeDirectionInfo{
			data:                 rd,
			shouldUpdatePeerAddr: v.shouldUpdatePeerAddr,
			peerAddr:             v.peerAddr,
			expires:              expires,
		}
	} else {
		c.values[key] = &routeDirectionInfo{
			data:    rd,
			expires: expires,
		}
	}
	c.mu.Unlock()

	return nil
}

// PutPeerAddr puts a peer's address into a cache. If a key and value exist, overwrite them.
func (c *routeDirectionCache) PutPeerAddr(key string, addr netip.AddrPort, expires int64) error {
	if !c.isReady.Load() {
		return fmt.Errorf("cache is not ready")
	}

	c.mu.Lock()
	if v, ok := c.values[key]; ok {
		c.values[key] = &routeDirectionInfo{
			data:                 v.data,
			shouldUpdatePeerAddr: true,
			peerAddr:             addr,
			expires:              expires,
		}
	} else {
		c.values[key] = &routeDirectionInfo{
			shouldUpdatePeerAddr: true,
			peerAddr:             addr,
			expires:              expires,
		}
	}
	c.mu.Unlock()

	return nil
}

// expired determines if it has expires.
func (d *routeDirectionInfo) expired(time int64) bool {
	if d.expires == 0 {
		return true
	}
	return time > d.expires
}
