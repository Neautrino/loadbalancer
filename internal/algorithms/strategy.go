package algorithms

import (
	"net/http"

	"github.com/neautrino/loadbalancer/internal/pool"
)

type Strategy interface {
	Next(healthy []*pool.Backend, r *http.Request) *pool.Backend 
}