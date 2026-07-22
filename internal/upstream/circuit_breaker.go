package upstream

import (
	"sync"
	"time"
)

type CircuitBreakerState string

const (
	StateClosed   CircuitBreakerState = "Closed"
	StateOpen     CircuitBreakerState = "Open"
	StateHalfOpen CircuitBreakerState = "HalfOpen"
)

type CircuitBreaker struct {
	mu             sync.RWMutex
	state          CircuitBreakerState
	failures       int
	errorThreshold int
	timeout        time.Duration
	lastOpenTime   time.Time
}

func NewCircuitBreaker(errorThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:          StateClosed,
		errorThreshold: errorThreshold,
		timeout:        timeout,
	}
}

func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// CanAttempt returns true if a request might be allowed.
// It does not mutate the state.
func (cb *CircuitBreaker) CanAttempt() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	if cb.state == StateOpen {
		return time.Since(cb.lastOpenTime) >= cb.timeout
	}
	return true // Closed or HalfOpen
}

// AllowRequest checks if a request is allowed and mutates state if it transitions to HalfOpen.
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateOpen {
		if time.Since(cb.lastOpenTime) >= cb.timeout {
			cb.state = StateHalfOpen
			return true
		}
		return false
	}
	if cb.state == StateHalfOpen {
		return false // Only one request allowed in HalfOpen
	}
	return true
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen {
		cb.state = StateOpen
		cb.lastOpenTime = time.Now()
		return
	}

	if cb.state == StateClosed {
		cb.failures++
		if cb.failures >= cb.errorThreshold {
			cb.state = StateOpen
			cb.lastOpenTime = time.Now()
		}
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen {
		cb.state = StateClosed
		cb.failures = 0
		return
	}

	if cb.state == StateClosed {
		cb.failures = 0
	}
}
