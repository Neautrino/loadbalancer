package internal

import (
	"log"
	"net/http"
)

type LoadBalancer struct {
	addr string
	pool *ServerPool
}

func NewLoadBalancer(addr string, backendURLs []string) (*LoadBalancer, error) {
	pool, err := NewServerPool(backendURLs)
	if err != nil {
		return nil, err
	}

	return &LoadBalancer{
		addr: addr,
		pool: pool,
	}, nil
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request){
	b := lb.pool.NextRoundRobin()
	if b == nil {
		http.Error(w, "no backend available", http.StatusServiceUnavailable)
		return 
	}

	log.Printf("[lb] %s %s -> %s", r.Method, r.URL.Path, b.URL)
	b.Proxy.ServeHTTP(w, r)
}

func (lb *LoadBalancer) Start() error {
	log.Printf("Load balancer listening on %s\n", lb.addr)
	return  http.ListenAndServe(lb.addr, lb)
}