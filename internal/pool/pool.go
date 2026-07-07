package pool

type ServerPool struct {
	backends []*Backend
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

func (p *ServerPool) Backends() []*Backend {
	return  p.backends
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