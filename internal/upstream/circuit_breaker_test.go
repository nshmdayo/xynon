package upstream

import (
	"testing"
	"time"
)

func TestCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond)

	// Should be Closed initially
	if cb.State() != StateClosed {
		t.Fatalf("expected state Closed, got %s", cb.State())
	}
	if !cb.CanAttempt() {
		t.Fatalf("CanAttempt should be true in Closed state")
	}

	// 1st failure
	cb.RecordFailure()
	if cb.State() != StateClosed {
		t.Fatalf("expected state Closed after 1 failure, got %s", cb.State())
	}

	// 2nd failure -> Open
	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Fatalf("expected state Open after 2 failures, got %s", cb.State())
	}
	if cb.CanAttempt() {
		t.Fatalf("CanAttempt should be false immediately after Open")
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// After timeout, CanAttempt should be true (ready for HalfOpen)
	if !cb.CanAttempt() {
		t.Fatalf("CanAttempt should be true after timeout")
	}

	// Acquiring the first request should succeed and transition to HalfOpen
	if !cb.AllowRequest() {
		t.Fatalf("AllowRequest should return true for first request after timeout")
	}
	if cb.State() != StateHalfOpen {
		t.Fatalf("expected state HalfOpen, got %s", cb.State())
	}

	// Acquiring a second request should fail (only one allowed in HalfOpen)
	if cb.AllowRequest() {
		t.Fatalf("AllowRequest should return false for concurrent requests in HalfOpen")
	}

	// The first request succeeds -> Closed
	cb.RecordSuccess()
	if cb.State() != StateClosed {
		t.Fatalf("expected state Closed after success in HalfOpen, got %s", cb.State())
	}
}

func TestCircuitBreaker_FailureInHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(1, 50*time.Millisecond)

	// 1st failure -> Open
	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Fatalf("expected state Open, got %s", cb.State())
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Acquire -> HalfOpen
	cb.AllowRequest()
	if cb.State() != StateHalfOpen {
		t.Fatalf("expected state HalfOpen, got %s", cb.State())
	}

	// Failure in HalfOpen -> Open
	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Fatalf("expected state Open after failure in HalfOpen, got %s", cb.State())
	}
	
	// Ensure CanAttempt is false immediately after it goes back to Open
	if cb.CanAttempt() {
		t.Fatalf("CanAttempt should be false after returning to Open")
	}
}
