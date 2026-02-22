package balancer

import (
	"sync"

	"github.com/mochivi/relay/internal/backend"
)

type LeastConnectionsBalancer struct {
	backends []*backend.Backend
	mux      sync.Mutex
}

func (b *LeastConnectionsBalancer) Next() *backend.Backend {
	b.mux.Lock()
	defer b.mux.Unlock()
	if len(b.backends) == 0 {
		return nil
	}

	selected := 0
	min := b.backends[0].Connections.Load()
	for i := 1; i < len(b.backends); i++ {
		c := b.backends[i].Connections.Load()
		if c < min {
			min = c
			selected = i
		}
	}

	b.backends[selected].Connections.Add(1)
	return b.backends[selected]
}

func (b *LeastConnectionsBalancer) Algorithm() string {
	return "least_connections"
}
