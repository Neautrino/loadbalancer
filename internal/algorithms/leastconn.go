package algorithms

import (
	"net/http"

	"github.com/neautrino/loadbalancer/internal/pool"
)

type LeastConn struct {}

func NewLeastConn() *LeastConn { 
	return &LeastConn{} 
}

func (lc *LeastConn) Next(healthy []*pool.Backend, r *http.Request) *pool.Backend {
	if len(healthy) == 0 {
		return  nil
	}
	best := healthy[0]
	for _, b := range healthy[1:] {
		if b.ActiveConns() < best.ActiveConns() {
			best = b
		}
	}

	return best
}