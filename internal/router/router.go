package router

import (
	"net/http"

	"github.com/mochivi/relay/internal/config"
	"github.com/mochivi/relay/internal/service"
)

type Router struct {
	tree     *tree
	services map[string]*service.Service
}

func NewRouter(routesCfg []*config.RouteConfig, services map[string]*service.Service) (*Router, error) {
	patterns := make([]string, 0, len(routesCfg))
	names := make([]string, 0, len(routesCfg))
	for _, route := range routesCfg {
		patterns = append(patterns, route.Pattern)
		names = append(names, route.Service)
	}

	tree, err := newTreeFromPatterns(patterns, names)
	if err != nil {
		return nil, err
	}
	tree.print()

	return &Router{
		tree:     tree,
		services: services,
	}, nil
}

func (r *Router) Match(req *http.Request) (*service.Service, bool) {
	serviceName, ok := r.tree.search(req.URL.Path)
	if !ok {
		return nil, false
	}
	service, ok := r.services[serviceName]
	if !ok {
		return nil, false
	}
	return service, true
}
