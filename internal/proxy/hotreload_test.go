package proxy

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nshmdayo/xynon/internal/config"
	pluginpkg "github.com/nshmdayo/xynon/internal/plugin"
)

const validWasm = "../plugin/testdata/valid/plugin.wasm"

func copyWasm(t *testing.T, dst string) {
	t.Helper()
	data, err := os.ReadFile(validWasm)
	if err != nil {
		t.Fatalf("read valid wasm: %v", err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatalf("write wasm: %v", err)
	}
}

func setupRuntime(t *testing.T) *pluginpkg.Runtime {
	t.Helper()
	ctx := context.Background()
	rt, err := pluginpkg.NewRuntime(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { rt.Close(context.Background()) })
	return rt
}

func TestHotReload_AddPlugin(t *testing.T) {
	dir := t.TempDir()
	wasmPath := filepath.Join(dir, "add-header.wasm")

	rt := setupRuntime(t)
	reg := &ChainRegistry{}
	reg.Store(NewChain(nil))

	entries := []pluginpkg.ChainEntry{{Name: "add-header", Type: config.PluginTypeWasm, Path: wasmPath}}
	cfg := WatcherConfig{
		PluginDir: dir,
		Entries:   entries,
		Limits:    pluginpkg.DefaultLimits(),
		Runtime:   rt,
		Registry:  reg,
	}

	w, err := NewWatcher(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	w.Start(ctx)
	defer w.Stop()

	// Drop the WASM file → should trigger reload.
	copyWasm(t, wasmPath)

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		chain := reg.Load()
		if chain != nil && len(chain.Plugins()) == 1 {
			return // success
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Error("hot-reload did not pick up new plugin within timeout")
}

func TestHotReload_InvalidFile_KeepsChain(t *testing.T) {
	dir := t.TempDir()
	wasmPath := filepath.Join(dir, "add-header.wasm")

	rt := setupRuntime(t)
	reg := &ChainRegistry{}
	original := NewChain(nil)
	reg.Store(original)

	entries := []pluginpkg.ChainEntry{{Name: "add-header", Type: config.PluginTypeWasm, Path: wasmPath}}
	cfg := WatcherConfig{
		PluginDir: dir,
		Entries:   entries,
		Limits:    pluginpkg.DefaultLimits(),
		Runtime:   rt,
		Registry:  reg,
	}

	w, err := NewWatcher(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	w.Start(ctx)
	defer w.Stop()

	// Write an invalid (non-WASM) file.
	if err := os.WriteFile(wasmPath, []byte("not a wasm file"), 0o644); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		current := reg.Load()
		if current != original {
			t.Errorf("expected original empty chain to be preserved, but chain was changed")
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func TestHotReload_ConfigFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	rt := setupRuntime(t)
	reg := &ChainRegistry{}
	reg.Store(NewChain(nil))

	reloadCalled := make(chan struct{}, 1)

	cfg := WatcherConfig{
		PluginDir:  dir,
		ConfigPath: configPath,
		ReloadConfig: func() ([]pluginpkg.ChainEntry, error) {
			select {
			case reloadCalled <- struct{}{}:
			default:
			}
			return nil, nil
		},
		Entries:   nil,
		Limits:    pluginpkg.DefaultLimits(),
		Runtime:   rt,
		Registry:  reg,
	}

	w, err := NewWatcher(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	w.Start(ctx)
	defer w.Stop()

	if err := os.WriteFile(configPath, []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-reloadCalled:
		// Success
	case <-time.After(3 * time.Second):
		t.Error("hot-reload did not call ReloadConfig within timeout")
	}
}

func TestHotReload_RPCPluginTrigger(t *testing.T) {
	dir := t.TempDir()
	rpcPath := filepath.Join(dir, "my-rpc-plugin")

	rt := setupRuntime(t)
	reg := &ChainRegistry{}
	reg.Store(NewChain(nil))

	reloadCalled := make(chan struct{}, 1)

	cfg := WatcherConfig{
		PluginDir: dir,
		ReloadConfig: func() ([]pluginpkg.ChainEntry, error) {
			select {
			case reloadCalled <- struct{}{}:
			default:
			}
			return nil, nil
		},
		Entries:   nil,
		Limits:    pluginpkg.DefaultLimits(),
		Runtime:   rt,
		Registry:  reg,
	}

	w, err := NewWatcher(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	w.Start(ctx)
	defer w.Stop()

	if err := os.WriteFile(rpcPath, []byte("executable binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	select {
	case <-reloadCalled:
		// Success
	case <-time.After(3 * time.Second):
		t.Error("hot-reload did not trigger on RPC plugin file change")
	}
}
