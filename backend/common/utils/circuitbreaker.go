package utils

import (
	"sync"
	"time"
)

// CircuitBreaker is a lightweight, in-memory circuit breaker implementation
// used to guard external dependencies (e.g. gRPC, MQ).
type CircuitBreaker struct {
	mu               sync.Mutex
	failures         int
	failureThreshold int
	openUntil        time.Time
	openDuration     time.Duration
}

// NewCircuitBreaker creates a new breaker with the given threshold and open duration.
// When consecutive failures reach failureThreshold, the breaker opens for openDuration.
func NewCircuitBreaker(threshold int, openDuration time.Duration) *CircuitBreaker {
	if threshold <= 0 {
		threshold = 5
	}
	if openDuration <= 0 {
		openDuration = 5 * time.Second
	}
	return &CircuitBreaker{
		failureThreshold: threshold,
		openDuration:     openDuration,
	}
}

// Allow reports whether a call is allowed. When the breaker is open and the
// cool-down period has not elapsed, it returns false.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	if now.After(cb.openUntil) {
		// Half-open / closed state: allow a trial request.
		return true
	}
	// Still within open window.
	return false
}

// OnSuccess should be called when a protected call succeeds.
// It resets the failure counter and closes the breaker.
func (cb *CircuitBreaker) OnSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.openUntil = time.Time{}
}

// OnFailure should be called when a protected call fails with a retriable error.
func (cb *CircuitBreaker) OnFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	if cb.failures >= cb.failureThreshold {
		cb.openUntil = time.Now().Add(cb.openDuration)
	}
}
