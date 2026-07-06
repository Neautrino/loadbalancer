package internal

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type LoadBalancer struct {
	addr string
	proxy *httputil.ReverseProxy
}

func NewLoadBalancer(addr string, backendURL string) (*LoadBalancer, error) {
	target, err := url.Parse(backendURL)
	if err != nil {
		return nil, err
	}

	return &LoadBalancer{
		addr: addr,
		proxy: httputil.NewSingleHostReverseProxy(target),
	}, nil
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request){
	lb.proxy.ServeHTTP(w, r)
}

func (lb *LoadBalancer) Start() error {
	log.Printf("Load balancer listening on %s\n", lb.addr)
	return  http.ListenAndServe(lb.addr, lb)
}