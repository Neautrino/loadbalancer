package algorithms

import (
	"net/http"
	"sync"

	"github.com/neautrino/loadbalancer/internal/pool"
)

type WeightedRoundRobin struct {
	mu sync.Mutex
}

func (w *WeightedRoundRobin) Next(healthy []*pool.Backend, r *http.Request) *pool.Backend {
	if len(healthy) == 0 {
		return nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()

	total := 0
	var best *pool.Backend
	
	for _, b := range healthy {
		b.CurrentWeight += b.Weight
		total += b.Weight
		if best == nil || b.CurrentWeight > best.CurrentWeight {
			best = b
		}
	}
	best.CurrentWeight -= total
	return  best
}