package pool

import (
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type Backend struct {
	URL           *url.URL
	Proxy         *httputil.ReverseProxy
	alive         atomic.Bool
	activeConns   atomic.Int64
	Weight        int
	CurrentWeight int
}

func NewBackend(rawUrl string) (*Backend, error) {
	target, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	b := &Backend{
		URL:   target,
		Proxy: httputil.NewSingleHostReverseProxy(target),
		Weight: 1,
	}
	b.alive.Store(true)
	return b, nil
}

func (b *Backend) IsAlive() bool {
	return b.alive.Load()
}

func (b *Backend) SetAlive(up bool) {
	b.alive.Store(up)
}

func (b *Backend) IncrConns() {
	b.activeConns.Add(1)
}

func (b *Backend) DecrConns() {
	b.activeConns.Add(-1)
}

func (b *Backend) ActiveConns() int64 {
	return b.activeConns.Load()
}