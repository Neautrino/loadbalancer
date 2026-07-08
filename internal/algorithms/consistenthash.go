package algorithms

import (
	"hash/fnv"
	"net/http"
	"sort"
	"strconv"

	"github.com/neautrino/loadbalancer/internal/pool"
)

type ConsistentHash struct {
	replicas int
}

func NewConsistentHash(replicas int) *ConsistentHash {
	return  &ConsistentHash{replicas: replicas}
}

func (c *ConsistentHash) Next(healthy []*pool.Backend, r *http.Request) *pool.Backend {
	if len(healthy) == 0 {
		return nil
	}

	type vnode struct {
		hash uint32
		backend *pool.Backend
	}
	ring := make([]vnode, 0, len(healthy)*c.replicas)
	for _, b := range healthy {
		for i := 0; i < c.replicas; i++ {
			ring = append(ring, vnode{
				hash: hashKey(b.URL.String() + ":" + strconv.Itoa(i)),
				backend: b,
			})
		}
	}

	sort.Slice(ring, func(i, j int) bool {
		return ring[i].hash < ring[j].hash
	})

	key := hashKey(clientIp(r))
	idx := sort.Search(len(ring), func (i int) bool {
		return ring[i].hash >= key
	})
	if idx == len(ring) {
		idx = 0
	}
	return ring[idx].backend
}

func hashKey(s string) uint32 {
	h := fnv.New32a()
	h.Write(([]byte(s)))
	return h.Sum32()
}