package pool

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL           *url.URL
	Proxy         *httputil.ReverseProxy
	Breaker *CircuitBreaker
	alive         atomic.Bool
	activeConns   atomic.Int64
	Weight        int
	CurrentWeight int
}

func NewBackend(rawUrl string, weight int, cbThreshold int, cbCooldown time.Duration) (*Backend, error) {
	target, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	b := &Backend{
		URL:   target,
		Breaker: NewCircuitBreaker(cbThreshold, cbCooldown),
		Weight: weight,
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.ModifyResponse = func(r *http.Response) error {
		b.Breaker.RecordSuccess()
		return nil
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		b.Breaker.RecordFailure()
		slog.Warn("proxy error", "url", b.URL.String(), "err", err)
		http.Error(w, "backend unavailable", http.StatusBadGateway)
	}

	b.Proxy = proxy
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

func (b *Backend) IsAvailable() bool {
	return  b.alive.Load() && b.Breaker.Allow()
}