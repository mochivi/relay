package proxy

import (
	"net/http"
)

func (p *Proxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	service, ok := p.router.Match(req)
	if !ok {
		http.NotFound(w, req)
		return
	}
	service.ServeNext(w, req)
}
