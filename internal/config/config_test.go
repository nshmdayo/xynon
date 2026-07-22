package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_Valid(t *testing.T) {
	path := writeTemp(t, `
listen: ":8080"
metrics:
  enabled: true
  address: ":9091"
plugins:
  dir: "/tmp/plugins"
  chain:
    - name: add-header
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Listen != ":8080" {
		t.Errorf("listen = %q, want :8080", cfg.Listen)
	}
	if cfg.Plugins.Dir != "/tmp/plugins" {
		t.Errorf("plugins.dir = %q, want /tmp/plugins", cfg.Plugins.Dir)
	}
	if cfg.Metrics.Enabled != true {
		t.Errorf("metrics.enabled = %v, want true", cfg.Metrics.Enabled)
	}
	if cfg.Metrics.Address != ":9091" {
		t.Errorf("metrics.address = %q, want :9091", cfg.Metrics.Address)
	}
	if len(cfg.Plugins.Chain) != 1 || cfg.Plugins.Chain[0].Name != "add-header" {
		t.Errorf("unexpected chain: %+v", cfg.Plugins.Chain)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_MissingListen(t *testing.T) {
	path := writeTemp(t, `
plugins:
  dir: "/tmp/plugins"
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing listen")
	}
}

func TestLoad_MissingPluginsDir(t *testing.T) {
	path := writeTemp(t, `
listen: ":8080"
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing plugins.dir")
	}
}

func TestLoad_MissingBoth(t *testing.T) {
	path := writeTemp(t, `{}`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing required fields")
	}
}

func TestSave_RoundTrip(t *testing.T) {
	path := writeTemp(t, `
listen: ":9090"
plugins:
  dir: "/tmp/p"
  chain:
    - name: foo
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	cfg.Plugins.Chain = append(cfg.Plugins.Chain, PluginEntry{Name: "bar"})
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	reloaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load after Save: %v", err)
	}
	if len(reloaded.Plugins.Chain) != 2 {
		t.Errorf("chain len = %d, want 2", len(reloaded.Plugins.Chain))
	}
	if reloaded.Plugins.Chain[1].Name != "bar" {
		t.Errorf("chain[1].name = %q, want bar", reloaded.Plugins.Chain[1].Name)
	}
}

func TestSave_RemoveFromChain(t *testing.T) {
	path := writeTemp(t, `
listen: ":9090"
plugins:
  dir: "/tmp/p"
  chain:
    - name: foo
    - name: bar
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	cfg.Plugins.Chain = cfg.Plugins.Chain[:1] // remove bar
	if err := Save(path, cfg); err != nil {
		t.Fatal(err)
	}

	reloaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(reloaded.Plugins.Chain) != 1 || reloaded.Plugins.Chain[0].Name != "foo" {
		t.Errorf("unexpected chain after remove: %+v", reloaded.Plugins.Chain)
	}
}
