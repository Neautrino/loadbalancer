package internal

import (
	"log"
	"net/http"
	"time"
)

type HealthChecker struct {
	pool *ServerPool
	interval time.Duration
	path string
	client *http.Client
}

func NewHealthChecker(pool *ServerPool, interval time.Duration, path string) *HealthChecker{
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
	for _, b := range hc.pool.backends {
		hc.check(b)
	}
}

func (hc *HealthChecker) check(b *Backend) {
	url := b.URL.String() + hc.path
	resp, err := hc.client.Get(url)
	if err != nil {
		b.SetAlive(false)
		log.Printf("*[health] %s DOWN (%v)", b.URL, err)
		return
	}
	defer resp.Body.Close()

	alive := resp.StatusCode == http.StatusOK
	b.SetAlive(alive)
	if !alive {
		log.Printf("[health] %s unhealthy (status %d)", b.URL, resp.StatusCode)
	}
 }