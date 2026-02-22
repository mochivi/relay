package balancer

import (
	"sync"

	"github.com/mochivi/relay/internal/backend"
)

type RoundRobinBalancer struct {
	backends []*backend.Backend
	mux      sync.Mutex
	index    int
}

func (b *RoundRobinBalancer) Next() *backend.Backend {
	b.mux.Lock()
	defer b.mux.Unlock()

	if len(b.backends) == 0 {
		return nil
	}
	if b.index == len(b.backends) {
		b.index = 0
	}
	backend := b.backends[b.index]
	b.index++
	backend.Connections.Add(1)
	return backend
}

func (b *RoundRobinBalancer) Algorithm() string {
	return "round_robin"
}
