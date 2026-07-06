package internal

import "sync/atomic"

type ServerPool struct {
	backends []*Backend
	counter atomic.Uint64
}

func NewServerPool(urls []string) (*ServerPool, error) {
	pool := &ServerPool{}
	for _, raw := range urls {
		b, err := NewBackend(raw)
		if err != nil {
			return nil, err
		}
		pool.backends = append(pool.backends, b)
	}

	return pool, nil
}

func (p *ServerPool) Healthy() []*Backend {
	healthy := make([]*Backend, 0, len(p.backends))
	for _, b := range p.backends {
		if b.IsAlive() {
			healthy = append(healthy, b)
		}
	}

	return  healthy
}

func (p *ServerPool) NextRoundRobin() *Backend {
	healthy := p.Healthy()
	if len(healthy) == 0 {
		return nil
	}

	n := p.counter.Add(1)
	return  healthy[(n-1)%uint64(len(healthy))]
}