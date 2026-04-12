package plugincli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuild_TinyGoNotFound(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "plugins")
	_ = os.MkdirAll(pluginDir, 0o755)
	cfgPath := writeCfg(t, dir, "listen: \":8080\"\nplugins:\n  dir: "+pluginDir+"\n  chain: []\n")

	t.Setenv("TINYGO", "/nonexistent/tinygo")
	err := runBuild([]string{"-config", cfgPath, "./someplugin"})
	if err == nil {
		t.Fatal("expected error when tinygo not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestBuild_TinyGoEnvVar(t *testing.T) {
	// Set TINYGO to a nonexistent path; resolveTinyGo should return an error
	// that mentions the TINYGO variable value (not a generic PATH error).
	t.Setenv("TINYGO", "/custom/path/tinygo")
	_, err := resolveTinyGo()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "/custom/path/tinygo") {
		t.Errorf("error should include TINYGO path, got: %v", err)
	}
}

func TestBuild_GOROOT_Env(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "plugins")
	_ = os.MkdirAll(pluginDir, 0o755)
	cfgPath := writeCfg(t, dir, "listen: \":8080\"\nplugins:\n  dir: "+pluginDir+"\n  chain: []\n")

	// Create a fake tinygo script that captures env.
	fakeScript := filepath.Join(dir, "tinygo")
	script := "#!/bin/sh\nenv | grep GOROOT > " + filepath.Join(dir, "env.txt") + "\nexit 0\n"
	if err := os.WriteFile(fakeScript, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("TINYGO", fakeScript)
	t.Setenv("XYNON_TINYGO_GOROOT", "/custom/goroot")

	srcDir := filepath.Join(dir, "myplugin")
	_ = os.MkdirAll(srcDir, 0o755)

	_ = runBuild([]string{"-config", cfgPath, srcDir})

	data, err := os.ReadFile(filepath.Join(dir, "env.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "GOROOT=/custom/goroot") {
		t.Errorf("GOROOT not passed to tinygo; captured env: %s", data)
	}
}
