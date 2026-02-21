package balancer

import "github.com/mochivi/relay/internal/backend"

type RoundRobinBalancer struct {
	backends []backend.Backend
	index    int
}

func (r *RoundRobinBalancer) Next() backend.Backend {
	if r.index == len(r.backends) {
		r.index = 0
	}
	backend := r.backends[r.index]
	r.index++
	return backend
}

func (r *RoundRobinBalancer) Algorithm() string {
	return "round_robin"
}
