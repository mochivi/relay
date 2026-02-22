package proxy

import (
	"net/http"
	"net/http/httputil"
)

func (p *Proxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	service, ok := p.router.Match(req)
	if !ok {
		http.NotFound(w, req)
		return
	}
	backend := service.Balancer.Next()
	if backend == nil {
		return
	}
	defer backend.Connections.Add(-1)

	revProxy := httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(backend.URL)
		},
	}
	revProxy.ServeHTTP(w, req)
}
