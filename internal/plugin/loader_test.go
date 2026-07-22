package plugin

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/nshmdayo/xynon/internal/config"
)

// TestLoadChain verifies loading a chain of plugins.
func TestLoadChain(t *testing.T) {
	// Build the dummy rpc plugin for testing
	tmpDir := t.TempDir()

	binPath := filepath.Join(tmpDir, "dummy_rpc")
	ctx, cancel := context.WithCancel(context.Background())
	if d, ok := t.Deadline(); ok {
		ctx, cancel = context.WithDeadline(context.Background(), d)
	}
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, "./testdata/dummy_rpc/main.go")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build dummy rpc plugin: %s\n%v", string(out), err)
	}

	rt, err := NewRuntime(context.Background())
	if err != nil {
		t.Fatalf("failed to create runtime: %v", err)
	}
	defer rt.Close(context.Background())

	entries := []ChainEntry{
		{
			Name: "my_rpc",
			Type: config.PluginTypeRPC,
			Path: binPath,
		},
		{
			Name: "my_wasm",
			Type: config.PluginTypeWasm,
			Path: "testdata/valid/plugin.wasm",
		},
	}

	handlers, err := LoadChain(context.Background(), rt, entries, DefaultLimits())
	if err != nil {
		t.Fatalf("LoadChain failed: %v", err)
	}

	if len(handlers) != 2 {
		t.Fatalf("expected 2 handlers, got %d", len(handlers))
	}

	// Clean up
	for _, h := range handlers {
		h.Close(context.Background())
	}
}

func TestLoadChain_UnknownType(t *testing.T) {
	rt, err := NewRuntime(context.Background())
	if err != nil {
		t.Fatalf("failed to create runtime: %v", err)
	}
	defer rt.Close(context.Background())

	entries := []ChainEntry{
		{
			Name: "unknown_plugin",
			Type: "invalid_type",
			Path: "some/path",
		},
	}

	_, err = LoadChain(context.Background(), rt, entries, DefaultLimits())
	if err == nil {
		t.Fatalf("expected error for unknown type, got nil")
	}
}

func TestLoadChain_ErrorCleanup(t *testing.T) {
	// Build the dummy rpc plugin for testing
	tmpDir, err := os.MkdirTemp("", "loader-test-cleanup")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binPath := filepath.Join(tmpDir, "dummy_rpc")
	ctx, cancel := context.WithCancel(context.Background())
	if d, ok := t.Deadline(); ok {
		ctx, cancel = context.WithDeadline(context.Background(), d)
	}
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, "./testdata/dummy_rpc/main.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build dummy rpc plugin: %v", err)
	}

	rt, err := NewRuntime(context.Background())
	if err != nil {
		t.Fatalf("failed to create runtime: %v", err)
	}
	defer rt.Close(context.Background())

	entries := []ChainEntry{
		{
			Name: "valid_rpc",
			Type: config.PluginTypeRPC,
			Path: binPath,
		},
		{
			Name: "invalid_wasm",
			Type: config.PluginTypeWasm,
			Path: "testdata/invalid/does_not_exist.wasm",
		},
	}

	handlers, err := LoadChain(context.Background(), rt, entries, DefaultLimits())
	if err == nil {
		t.Fatalf("expected error for missing wasm, got nil")
	}
	if handlers != nil {
		t.Fatalf("expected nil handlers on error, got %v", handlers)
	}
}

func TestDefaultLimits(t *testing.T) {
	limits := DefaultLimits()
	if limits.Timeout != 5*time.Second {
		t.Errorf("expected 5s timeout, got %v", limits.Timeout)
	}
	if limits.MemoryMB != 64 {
		t.Errorf("expected 64MB memory, got %v", limits.MemoryMB)
	}
}
