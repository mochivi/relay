package service

import (
	"github.com/mochivi/relay/internal/backend"
	"github.com/mochivi/relay/internal/balancer"
	"github.com/mochivi/relay/internal/config"
)

type Service struct {
	Name     string
	Balancer balancer.Balancer
}

func NewService(cfg config.ServiceConfig) (*Service, error) {
	backends := make([]backend.Backend, 0, len(cfg.Backends))
	for _, rawBackend := range cfg.Backends {
		backend, err := backend.NewBackend(rawBackend, 1)
		if err != nil {
			return nil, err
		}
		backends = append(backends, backend)
	}

	balancer, err := balancer.NewBalancer(cfg.Algorithm, backends)
	if err != nil {
		return nil, err
	}

	return &Service{
		Name:     cfg.Name,
		Balancer: balancer,
	}, nil
}
