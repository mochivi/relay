package balancer

import (
	"errors"

	"github.com/mochivi/relay/internal/backend"
)

type Balancer interface {
	Next() backend.Backend
	Algorithm() string
}

func NewBalancer(algorithm string, backends []backend.Backend) (Balancer, error) {
	switch algorithm {
	case "round_robin":
		return &RoundRobinBalancer{backends: backends}, nil
	}
	return nil, errors.New("algorithm not supported")
}
