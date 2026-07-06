package internal

import (
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type Backend struct {
	URL *url.URL
	Proxy *httputil.ReverseProxy
	alive atomic.Bool
}

func NewBackend(rawUrl string) (*Backend, error) {
	target, err := url.Parse(rawUrl)
	if err != nil {
		return  nil, err
	}

	b := &Backend{
		URL: target,
		Proxy: httputil.NewSingleHostReverseProxy(target),
	}
	b.alive.Store(true)
	return  b, nil
}

func (b* Backend) IsAlive() bool {
	return b.alive.Load()
}

func (b *Backend) SetAlive(up bool) {
	b.alive.Store(up)
}