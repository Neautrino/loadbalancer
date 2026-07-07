package algorithms

import (
	"net/http"
	"sync/atomic"

	"github.com/neautrino/loadbalancer/internal/pool"
)

type Strategy interface {
	Next(healthy []*pool.Backend, r *http.Request) *pool.Backend 
}

type RoundRobin struct {
	counter atomic.Uint64
}

func (rr *RoundRobin) Next(healthy []*pool.Backend, r *http.Request) *pool.Backend {
	if len(healthy) == 0 {
		return  nil
	}
	n := rr.counter.Add(1)
	return healthy[(n-1)%uint64(len(healthy))]
}