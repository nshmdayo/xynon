package plugincli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nshmdayo/xynon/internal/config"
)

func setupRemoveEnv(t *testing.T, chain string) (cfgPath, pluginDir string) {
	t.Helper()
	dir := t.TempDir()
	pluginDir = filepath.Join(dir, "plugins")
	_ = os.MkdirAll(pluginDir, 0o755)
	cfgPath = writeCfg(t, dir, "listen: \":8080\"\nplugins:\n  dir: "+pluginDir+"\n  chain:\n"+chain)
	return
}

func TestRemove_Registered(t *testing.T) {
	cfgPath, pluginDir := setupRemoveEnv(t, "    - name: foo\n    - name: bar\n")
	// create dummy wasm so list works, but file should survive remove
	_ = os.WriteFile(filepath.Join(pluginDir, "foo.wasm"), []byte("x"), 0o644)

	if err := runRemove([]string{"-config", cfgPath, "foo"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, _ := config.Load(cfgPath)
	if len(cfg.Plugins.Chain) != 1 || cfg.Plugins.Chain[0].Name != "bar" {
		t.Errorf("unexpected chain after remove: %+v", cfg.Plugins.Chain)
	}

	// WASM file must still exist
	if _, err := os.Stat(filepath.Join(pluginDir, "foo.wasm")); err != nil {
		t.Errorf("wasm file should not be deleted: %v", err)
	}
}

func TestRemove_NotRegistered(t *testing.T) {
	cfgPath, _ := setupRemoveEnv(t, "    - name: foo\n")

	err := runRemove([]string{"-config", cfgPath, "nonexistent"})
	if err == nil {
		t.Fatal("expected error for unregistered plugin")
	}

	// config must be unchanged
	cfg, _ := config.Load(cfgPath)
	if len(cfg.Plugins.Chain) != 1 {
		t.Errorf("config should be unchanged, got chain: %+v", cfg.Plugins.Chain)
	}
}
