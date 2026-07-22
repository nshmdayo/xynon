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

	// wait for some ticks
	time.Sleep(50 * time.Millisecond)

	// healthyCount == 2 should be hit, status should be healthy, then it will hit unhealthyCount
	// max fails is 2, so it might need a bit more time to reach MaxFails and become unhealthy
	time.Sleep(100 * time.Millisecond)

	if srv.IsHealthy() {
		t.Errorf("expected server to be unhealthy after multiple 500 responses")
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
	
	// Start with context already canceled so it returns immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	
	hc.Start(ctx)
}
