package backend

import (
	"fmt"
	"net/url"
)

type Backend struct {
	URL    *url.URL
	Weight int
}

func NewBackend(rawUrl string, weight int) (Backend, error) {
	url, err := url.Parse(rawUrl)
	if err != nil {
		return Backend{}, fmt.Errorf("failed to parse URL: %w", err)
	}
	return Backend{
		URL:    url,
		Weight: weight,
	}, nil
}
