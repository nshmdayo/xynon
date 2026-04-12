package plugincli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nshmdayo/xynon/internal/config"
)

func setupAddEnv(t *testing.T) (cfgPath, pluginDir, wasmSrc string) {
	t.Helper()
	dir := t.TempDir()
	pluginDir = filepath.Join(dir, "plugins")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// dummy wasm source
	wasmSrc = filepath.Join(dir, "new-plugin.wasm")
	if err := os.WriteFile(wasmSrc, []byte("fakewasm"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfgPath = writeCfg(t, dir, "listen: \":8080\"\nplugins:\n  dir: "+pluginDir+"\n  chain: []\n")
	return
}

func TestAdd_NewPlugin(t *testing.T) {
	cfgPath, pluginDir, wasmSrc := setupAddEnv(t)

	if err := runAdd([]string{"-config", cfgPath, wasmSrc}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// file should be copied
	if _, err := os.Stat(filepath.Join(pluginDir, "new-plugin.wasm")); err != nil {
		t.Errorf("wasm not copied: %v", err)
	}

	// chain should contain the new plugin
	cfg, _ := config.Load(cfgPath)
	if len(cfg.Plugins.Chain) != 1 || cfg.Plugins.Chain[0].Name != "new-plugin" {
		t.Errorf("unexpected chain: %+v", cfg.Plugins.Chain)
	}
}

func TestAdd_DuplicateNotAddedToChain(t *testing.T) {
	cfgPath, _, wasmSrc := setupAddEnv(t)

	// Add once
	if err := runAdd([]string{"-config", cfgPath, wasmSrc}); err != nil {
		t.Fatal(err)
	}
	// Add again
	if err := runAdd([]string{"-config", cfgPath, wasmSrc}); err != nil {
		t.Fatal(err)
	}

	cfg, _ := config.Load(cfgPath)
	if len(cfg.Plugins.Chain) != 1 {
		t.Errorf("chain len = %d, want 1 (no duplicate)", len(cfg.Plugins.Chain))
	}
}

func TestAdd_MissingSourceFile(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "plugins")
	_ = os.MkdirAll(pluginDir, 0o755)
	cfgPath := writeCfg(t, dir, "listen: \":8080\"\nplugins:\n  dir: "+pluginDir+"\n  chain: []\n")

	err := runAdd([]string{"-config", cfgPath, "/nonexistent/path.wasm"})
	if err == nil {
		t.Fatal("expected error for missing source file")
	}

	// config must be unchanged
	cfg, _ := config.Load(cfgPath)
	if len(cfg.Plugins.Chain) != 0 {
		t.Errorf("config should be unchanged, got chain: %+v", cfg.Plugins.Chain)
	}
}
