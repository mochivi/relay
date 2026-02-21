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
	next := service.Balancer.Next()

	revProxy := httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(next.URL)
		},
	}
	revProxy.ServeHTTP(w, req)
}
