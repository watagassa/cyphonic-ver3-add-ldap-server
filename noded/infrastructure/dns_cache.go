package infrastructure

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
	"github.com/Pluslab/cyphonic/noded/usecase/repository"
)

var _ repository.DNSCache = (*dnsCache)(nil)

type dnsCache struct {
	isReady atomic.Bool
	bufs    map[uint32]*dnsBuf
	mu      sync.Mutex
}

// dnsBuf is the data to be cached.
type dnsBuf struct {
	port          uint16
	transactionID uint16
	expires       int64
}

func NewDNSCache() repository.DNSCache {
	return &dnsCache{
		isReady: atomic.Bool{},
		bufs:    make(map[uint32]*dnsBuf),
	}
}

func (c *dnsCache) Run(ctx context.Context) error {
	config.LogDebug("Starting DNS cache")
	defer config.LogDebug("Finished DNS cache")

	t := time.NewTicker(time.Second)
	defer t.Stop()

	c.isReady.Store(true)
	defer c.isReady.Store(false)

	for {
		select {
		case <-t.C:
			c.mu.Lock()
			for k, v := range c.bufs {
				if v.expired(time.Now().UnixNano()) {
					delete(c.bufs, k)
				}
			}
			c.mu.Unlock()
		case <-ctx.Done():
			return nil
		}
	}
}

// Get gets a value from a cache.
func (c *dnsCache) Get(key uint32) (port, transactionID uint16) {
	if !c.isReady.Load() {
		return 0, 0
	}

	c.mu.Lock()
	if v, ok := c.bufs[key]; ok {
		port = v.port
		transactionID = v.transactionID
	}
	c.mu.Unlock()

	return port, transactionID
}

// Put puts a value to a cache. If a key and value exists, overwrite it.
func (c *dnsCache) Put(key uint32, port, transactionID uint16, expires int64) {
	if !c.isReady.Load() {
		return
	}

	c.mu.Lock()
	if _, ok := c.bufs[key]; !ok {
		c.bufs[key] = &dnsBuf{
			port:          port,
			transactionID: transactionID,
			expires:       expires,
		}
	}
	c.mu.Unlock()
}

// expired determines if it has expires.
func (d *dnsBuf) expired(time int64) bool {
	if d.expires == 0 {
		return true
	}
	return time > d.expires
}
