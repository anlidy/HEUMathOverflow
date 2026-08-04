package utils

import (
	"testing"
	"time"
)

func TestNewCircuitBreaker_Defaults(t *testing.T) {
	cb := NewCircuitBreaker(0, 0)
	if cb.failureThreshold != 5 {
		t.Errorf("expected default threshold=5, got %d", cb.failureThreshold)
	}
	if cb.openDuration != 5*time.Second {
		t.Errorf("expected default openDuration=5s, got %v", cb.openDuration)
	}
}

func TestCircuitBreaker_InitiallyClosed(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Second)
	if !cb.Allow() {
		t.Error("new breaker should allow requests (closed state)")
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker(3, 50*time.Millisecond)
	cb.OnFailure()
	cb.OnFailure()
	if !cb.Allow() {
		t.Error("should still allow before threshold")
	}
	cb.OnFailure() // 3rd failure hits threshold
	if cb.Allow() {
		t.Error("breaker should be open after threshold reached")
	}
}

func TestCircuitBreaker_ReopensAfterDuration(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond)
	cb.OnFailure()
	cb.OnFailure()
	if cb.Allow() {
		t.Error("breaker should be open")
	}
	time.Sleep(60 * time.Millisecond)
	if !cb.Allow() {
		t.Error("breaker should allow after cool-down")
	}
}

func TestCircuitBreaker_ResetsOnSuccess(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Second)
	cb.OnFailure()
	cb.OnFailure()
	cb.OnSuccess()
	cb.OnFailure()
	cb.OnFailure()
	if !cb.Allow() {
		t.Error("failure count should have been reset by OnSuccess, should still allow")
	}
}
