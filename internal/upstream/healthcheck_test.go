package upstream

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nshmdayo/xynon/internal/config"
)

func TestActiveHealthChecker(t *testing.T) {
	healthyCount := 0
	unhealthyCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			if healthyCount < 2 {
				healthyCount++
				w.WriteHeader(http.StatusOK)
				return
			}
			unhealthyCount++
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	srv, _ := NewServer(ts.URL, nil)

	cfg := config.ActiveCheckConfig{
		Enabled:        true,
		Path:           "/health",
		ExpectedStatus: 200,
		Interval:       "10ms",
		Timeout:        "1s",
		MaxFails:       2,
	}

	hc := NewActiveHealthChecker(srv, cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hc.Start(ctx)

	// wait until it becomes unhealthy or timeout expires
	deadline := time.Now().Add(500 * time.Millisecond)
	for srv.IsHealthy() {
		if time.Now().After(deadline) {
			t.Errorf("expected server to be unhealthy after multiple 500 responses, but timeout reached")
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	cancel() // stop ticker
	time.Sleep(20 * time.Millisecond)
}

func TestActiveHealthChecker_InvalidConfigFallback(t *testing.T) {
	// test fallbacks
	cfg := config.ActiveCheckConfig{
		Interval: "invalid",
		Timeout:  "invalid",
	}
	srv, _ := NewServer("http://localhost:8080", nil)
	hc := NewActiveHealthChecker(srv, cfg)
	if hc.client.Timeout != 2*time.Second {
		t.Errorf("expected fallback timeout 2s, got %v", hc.client.Timeout)
	}
	
	// Start with context already canceled so it returns immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	
	hc.Start(ctx)
}
