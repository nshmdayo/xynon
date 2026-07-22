package upstream

import (
	"testing"
	"time"
)

func TestServer_RecordPassiveFailureAndSuccess(t *testing.T) {
	s, err := NewServer("http://localhost:8080", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !s.IsHealthy() {
		t.Errorf("expected server to be healthy initially")
	}

	failTimeout := 1 * time.Second

	// Record one failure
	s.RecordPassiveFailure(2, failTimeout)
	if !s.IsHealthy() {
		t.Errorf("expected server to still be healthy after 1 fail (max 2)")
	}

	// Record second failure -> should become unhealthy
	s.RecordPassiveFailure(2, failTimeout)
	if s.IsHealthy() {
		t.Errorf("expected server to be unhealthy after 2 fails")
	}
	
	if s.Status() != StatusUnhealthy {
		t.Errorf("expected StatusUnhealthy, got %v", s.Status())
	}

	// Record success -> should become healthy again
	s.RecordPassiveSuccess(failTimeout)
	if !s.IsHealthy() {
		t.Errorf("expected server to be healthy after success")
	}
	if s.Status() != StatusHealthy {
		t.Errorf("expected StatusHealthy, got %v", s.Status())
	}
}

func TestServer_AcquireCBAndRecordCB(t *testing.T) {
	cb := NewCircuitBreaker(1, 1*time.Minute)
	s, err := NewServer("http://localhost:8080", cb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !s.AcquireCB() {
		t.Errorf("expected to acquire cb initially")
	}

	s.RecordCBFailure()
	if s.AcquireCB() {
		t.Errorf("expected not to acquire cb after failure threshold reached")
	}

	// Test nil CB
	s2, _ := NewServer("http://localhost:8081", nil)
	if !s2.AcquireCB() {
		t.Errorf("expected to acquire cb when cb is nil")
	}
	s2.RecordCBFailure()
	s2.RecordCBSuccess()
}

func TestServer_Connections(t *testing.T) {
	s, _ := NewServer("http://localhost:8080", nil)
	s.IncConn()
	if s.ActiveConns() != 1 {
		t.Errorf("expected 1 active conn, got %d", s.ActiveConns())
	}
	s.DecConn()
	if s.ActiveConns() != 0 {
		t.Errorf("expected 0 active conns, got %d", s.ActiveConns())
	}
}
