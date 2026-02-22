package backend

import (
	"fmt"
	"net/url"
	"sync/atomic"
)

type Backend struct {
	URL         *url.URL
	Weight      int
	Connections atomic.Int64
}

func NewBackend(rawUrl string, weight int) (*Backend, error) {
	url, err := url.Parse(rawUrl)
	if err != nil {
		return &Backend{}, fmt.Errorf("failed to parse URL: %w", err)
	}
	return &Backend{
		URL:    url,
		Weight: weight,
	}, nil
}
