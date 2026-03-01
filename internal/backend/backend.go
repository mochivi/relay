package backend

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type Backend struct {
	URL         *url.URL
	Weight      int
	Connections atomic.Int64
	revProxy    *httputil.ReverseProxy
}

func NewBackend(rawUrl string, weight int) (*Backend, error) {
	url, err := url.Parse(rawUrl)
	if err != nil {
		return &Backend{}, fmt.Errorf("failed to parse URL: %w", err)
	}
	return &Backend{
		URL:      url,
		Weight:   weight,
		revProxy: httputil.NewSingleHostReverseProxy(url),
	}, nil
}

func (b *Backend) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	b.revProxy.ServeHTTP(w, req)
}
