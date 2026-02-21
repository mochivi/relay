package proxy

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/mochivi/relay/internal/config"
	"github.com/mochivi/relay/internal/router"
)

type Proxy struct {
	server *http.Server
	router *router.Router
}

func NewProxy(cfg config.GlobalConfig, router *router.Router) *Proxy {
	proxy := &Proxy{router: router}
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Addr, strconv.Itoa(cfg.Port)),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      proxy,
	}
	proxy.server = server

	return proxy
}

func (p *Proxy) Start() error {
	if err := p.server.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func (p *Proxy) Shutdown(ctx context.Context) error {
	if err := p.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}
	return nil
}
