package pool

type ServerPool struct {
	backends []*Backend
}

func NewServerPool(backends []*Backend) *ServerPool {
	return &ServerPool{backends: backends}
}

func (p *ServerPool) Backends() []*Backend {
	return  p.backends
}

func (p *ServerPool) Healthy() []*Backend {
	healthy := make([]*Backend, 0, len(p.backends))
	for _, b := range p.backends {
		if b.IsAvailable() {
			healthy = append(healthy, b)
		}
	}

	return  healthy
}