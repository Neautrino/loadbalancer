package internal

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/neautrino/loadbalancer/internal/pool"
)

type HealthChecker struct {
	pool *pool.ServerPool
	interval time.Duration
	path string
	client *http.Client
}

func NewHealthChecker(pool *pool.ServerPool, interval time.Duration, path string) *HealthChecker{
	return &HealthChecker{
		pool: pool,
		interval: interval,
		path: path,
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

func (hc *HealthChecker) Start() {
	go func() {
		hc.checkAll()
		ticker := time.NewTicker(hc.interval)
		defer ticker.Stop()
		for range ticker.C {
			hc.checkAll()
		}
	} ()
}

func (hc *HealthChecker) checkAll() {
	for _, b := range hc.pool.Backends() {
		hc.check(b)
	}
}

func (hc *HealthChecker) check(b *pool.Backend) {
	url := b.URL.String() + hc.path
	resp, err := hc.client.Get(url)
	if err != nil {
		b.SetAlive(false)
		slog.Error("backend down", "url", b.URL.String(), "err", err)
		return
	}
	defer resp.Body.Close()

	alive := resp.StatusCode == http.StatusOK
	b.SetAlive(alive)
	if !alive {
		slog.Warn("backend unhealthy", "url", b.URL.String(), "status", resp.StatusCode)
	}
 }