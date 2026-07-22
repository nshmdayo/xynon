package upstream

import (
	"context"
	"testing"

	"github.com/nshmdayo/xynon/internal/config"
)

func TestNewManagerAndGet(t *testing.T) {
	cfg := []config.UpstreamConfig{
		{
			Name: "backend1",
			Servers: []config.ServerConfig{
				{URL: "http://localhost:8081"},
			},
			Algorithm: "round_robin",
			CircuitBreaker: config.CircuitBreakerConfig{
				Enabled:        true,
				ErrorThreshold: 2,
				Timeout:        "1s",
			},
		},
		{
			Name: "backend2",
			Servers: []config.ServerConfig{
				{URL: "http://localhost:8082"},
			},
			Algorithm: "invalid", // should fallback to round_robin
		},
	}

	ctx := context.Background()
	m, err := NewManager(ctx, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	up1 := m.Get("backend1")
	if up1 == nil {
		t.Fatalf("expected to find backend1")
	}
	if up1.Name != "backend1" {
		t.Errorf("expected name backend1, got %s", up1.Name)
	}
	if up1.Config().Name != "backend1" {
		t.Errorf("expected config name backend1")
	}

	up2 := m.Get("backend2")
	if up2 == nil {
		t.Fatalf("expected to find backend2")
	}

	// Test Statuses
	statuses := m.Statuses()
	if len(statuses) != 2 {
		t.Errorf("expected 2 statuses, got %d", len(statuses))
	}

	// Test StopAll
	m.StopAll()
}

func TestNewManager_InvalidURL(t *testing.T) {
	cfg := []config.UpstreamConfig{
		{
			Name: "backend1",
			Servers: []config.ServerConfig{
				{URL: ":invalid-url"},
			},
		},
	}
	_, err := NewManager(context.Background(), cfg)
	if err == nil {
		t.Errorf("expected error for invalid URL")
	}
}
