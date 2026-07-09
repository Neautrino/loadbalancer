package internal

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/neautrino/loadbalancer/internal/algorithms"
	"github.com/neautrino/loadbalancer/internal/config"
	"github.com/neautrino/loadbalancer/internal/pool"
)

type LoadBalancer struct {
	addr string
	pool *pool.ServerPool
	strategy algorithms.Strategy
	healthInterval time.Duration
	healthPath string
}

func NewLoadBalancer(cfg *config.Config, strategy algorithms.Strategy) (*LoadBalancer, error) {
	var backends []*pool.Backend
	for _, bc := range cfg.Backends {
		b, err := pool.NewBackend(bc.URL, bc.Weight, cfg.CircuitBreaker.Threshold, time.Duration(cfg.CircuitBreaker.Cooldown))
		if err != nil {
			return nil , err
		}
		backends = append(backends, b)
	}

	return &LoadBalancer{
		addr: fmt.Sprintf(":%d", cfg.Port ),
		pool: pool.NewServerPool(backends),
		strategy: strategy,
		healthInterval: time.Duration(cfg.Health.Interval),
		healthPath: cfg.Health.Path,
	}, nil
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request){
	b := lb.strategy.Next(lb.pool.Healthy(), r)
	if b == nil {
		http.Error(w, "no backend available", http.StatusServiceUnavailable)
		return 
	}

	slog.Info("request", "method", r.Method, "path", r.URL.Path, "backend", b.URL.String())

	b.IncrConns()
	defer b.DecrConns()
	b.Proxy.ServeHTTP(w, r)
}

func (lb *LoadBalancer) Start() error {
	checker := NewHealthChecker(lb.pool, lb.healthInterval, lb.healthPath)
	checker.Start()

	handler := LoggingMiddleware(lb)

	slog.Info("Load balancer listening on", "addr", lb.addr)
	return  http.ListenAndServe(lb.addr, handler)
}