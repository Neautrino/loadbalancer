package pool

import (
	"log/slog"
	"sync"
	"time"
)

type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

type CircuitBreaker struct {
	mu sync.Mutex
	state State
	failures int
	threshold int
	cooldown time.Duration
	openedAt time.Time
}

func NewCircuitBreaker(threshold int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state: Closed,
		threshold: threshold,
		cooldown: cooldown,
	}
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state{
		case Open:
			if time.Since(cb.openedAt) > cb.cooldown {
				cb.state = HalfOpen
				slog.Info("circuit HALF-OPEN")
				return true
			}
			return  false
		default:
			return true
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	cb.state = Closed
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	if cb.state == HalfOpen || cb.failures >= cb.threshold {
		cb.state = Open
		cb.openedAt = time.Now()
		slog.Warn("circuit OPEN", "failures", cb.failures)
	}
}