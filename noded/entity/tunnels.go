package entity

import "sync"

type Tunnels struct {
	sync.RWMutex
	m map[PathID]Tunnel
}

func NewTunnels() *Tunnels {
	return &Tunnels{
		m: make(map[PathID]Tunnel),
	}
}

func (ts *Tunnels) Get(pathID PathID) (Tunnel, bool) {
	ts.RLock()
	defer ts.RUnlock()

	peer, ok := ts.m[pathID]
	return peer, ok
}

func (ts *Tunnels) Set(pathID PathID, tunnel Tunnel) {
	ts.Lock()
	defer ts.Unlock()

	ts.m[pathID] = tunnel
}

func (ts *Tunnels) GetRandom() (Tunnel, bool) {
	ts.RLock()
	defer ts.RUnlock()

	for _, t := range ts.m {
		return t, true
	}
	return nil, false
}
