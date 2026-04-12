package plugincli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeCfg(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestList_WithPlugins(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "plugins")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// create a dummy .wasm file
	if err := os.WriteFile(filepath.Join(pluginDir, "add-header.wasm"), []byte("fake"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfgPath := writeCfg(t, dir, "listen: \":8080\"\nplugins:\n  dir: "+pluginDir+"\n  chain:\n    - name: add-header\n")

	var buf bytes.Buffer
	if err := runList([]string{"-config", cfgPath}, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "[1]") || !strings.Contains(out, "add-header") {
		t.Errorf("expected chain position in output, got: %s", out)
	}
}

func TestList_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "plugins")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgPath := writeCfg(t, dir, "listen: \":8080\"\nplugins:\n  dir: "+pluginDir+"\n  chain: []\n")

	var buf bytes.Buffer
	if err := runList([]string{"-config", cfgPath}, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "no plugins found") {
		t.Errorf("expected 'no plugins found', got: %s", buf.String())
	}
}
