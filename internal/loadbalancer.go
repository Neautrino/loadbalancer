package internal

import (
	"log"
	"net/http"
	"time"

	"github.com/neautrino/loadbalancer/internal/algorithms"
	"github.com/neautrino/loadbalancer/internal/pool"
)

type LoadBalancer struct {
	addr string
	pool *pool.ServerPool
	strategy algorithms.Strategy
}

func NewLoadBalancer(addr string, backendURLs []string, strategy algorithms.Strategy) (*LoadBalancer, error) {
	pool, err := pool.NewServerPool(backendURLs)
	if err != nil {
		return nil, err
	}

	return &LoadBalancer{
		addr: addr,
		pool: pool,
		strategy: strategy,
	}, nil
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request){
	b := lb.strategy.Next(lb.pool.Healthy(), r)
	if b == nil {
		http.Error(w, "no backend available", http.StatusServiceUnavailable)
		return 
	}

	log.Printf("[lb] %s %s -> %s", r.Method, r.URL.Path, b.URL)
	b.Proxy.ServeHTTP(w, r)
}

func (lb *LoadBalancer) Start() error {
	checker := NewHealthChecker(lb.pool, 10 * time.Second, "/health")
	checker.Start()

	log.Printf("Load balancer listening on %s\n", lb.addr)
	return  http.ListenAndServe(lb.addr, lb)
}