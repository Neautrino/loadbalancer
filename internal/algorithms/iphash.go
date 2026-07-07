package algorithms

import (
	"hash/fnv"
	"net"
	"net/http"
	"strings"

	"github.com/neautrino/loadbalancer/internal/pool"
)

type IpHash struct{}

func (h *IpHash) Next(healthy []*pool.Backend, r *http.Request) *pool.Backend {
	if len(healthy) == 0 {
		return nil
	}

	ip := clientIp(r)
	hasher := fnv.New32a()
	hasher.Write([]byte(ip))
	idx := hasher.Sum32() % uint32(len(healthy))
	return  healthy[idx]
}

func clientIp(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return  r.RemoteAddr
	}
	return host
}