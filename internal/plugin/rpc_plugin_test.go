package plugin

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"net/rpc"
	hplugin "github.com/hashicorp/go-plugin"

	"github.com/nshmdayo/xynon/internal/plugin/abi"
)

func TestLoadRpcPlugin(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rpc-plugin-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binPath := filepath.Join(tmpDir, "dummy_rpc")
	cmd := exec.Command("go", "build", "-o", binPath, "./testdata/dummy_rpc/main.go")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build dummy rpc plugin: %s\n%v", string(out), err)
	}

	handler, err := LoadRpcPlugin(context.Background(), "dummy_rpc", binPath)
	if err != nil {
		t.Fatalf("LoadRpcPlugin failed: %v", err)
	}
	defer handler.Close(context.Background())

	if handler.Name() != "dummy_rpc" {
		t.Errorf("expected name dummy_rpc, got %s", handler.Name())
	}

	// Test OnRequest success
	header := make(http.Header)
	header.Set("X-Test", "1")
	action, sc, scStatus, err := handler.OnRequest(context.Background(), header)
	if err != nil {
		t.Fatalf("OnRequest failed: %v", err)
	}
	if action != abi.ActionContinue {
		t.Errorf("expected ActionContinue, got %v", action)
	}
	if sc {
		t.Errorf("expected no short circuit")
	}
	if header.Get("X-Dummy-Request") != "handled" {
		t.Errorf("header not updated: %v", header)
	}
	if header.Get("X-Test") != "1" {
		t.Errorf("original header lost: %v", header)
	}

	// Test OnRequest short circuit
	header = make(http.Header)
	header.Set("X-Short", "true")
	action, sc, scStatus, err = handler.OnRequest(context.Background(), header)
	if err != nil {
		t.Fatalf("OnRequest short circuit failed: %v", err)
	}
	if !sc {
		t.Errorf("expected short circuit")
	}
	if scStatus != 403 {
		t.Errorf("expected status 403, got %d", scStatus)
	}

	// Test OnRequest error
	header = make(http.Header)
	header.Set("X-Fail", "true")
	_, _, _, err = handler.OnRequest(context.Background(), header)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	// Test OnResponse success
	header = make(http.Header)
	header.Set("X-Test", "2")
	action, sc, scStatus, err = handler.OnResponse(context.Background(), header, 200)
	if err != nil {
		t.Fatalf("OnResponse failed: %v", err)
	}
	if action != abi.ActionContinue {
		t.Errorf("expected ActionContinue, got %v", action)
	}
	if header.Get("X-Dummy-Response") != "handled" {
		t.Errorf("header not updated: %v", header)
	}
	if header.Get("X-Test") != "2" {
		t.Errorf("original header lost: %v", header)
	}

	// Test OnResponse error
	header = make(http.Header)
	header.Set("X-Fail", "true")
	_, _, _, err = handler.OnResponse(context.Background(), header, 200)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestLoadRpcPlugin_InvalidPath(t *testing.T) {
	_, err := LoadRpcPlugin(context.Background(), "dummy", "/invalid/path/that/does/not/exist")
	if err == nil {
		t.Errorf("expected error for invalid path, got nil")
	}
}

// Test coverage for Server / Client implementation directly
type MockHandler struct{}
func (m *MockHandler) Name() string { return "mock" }
func (m *MockHandler) OnRequest(ctx context.Context, header http.Header) (abi.Action, bool, int, error) {
	return abi.ActionContinue, false, 0, nil
}
func (m *MockHandler) OnResponse(ctx context.Context, header http.Header, statusCode int) (abi.Action, bool, int, error) {
	return abi.ActionContinue, false, 0, nil
}
func (m *MockHandler) Close(ctx context.Context) error { return nil }

func TestXynonPlugin(t *testing.T) {
	p := &XynonPlugin{Impl: &MockHandler{}}
	
	srv, err := p.Server(&hplugin.MuxBroker{})
	if err != nil {
		t.Fatalf("Server() failed: %v", err)
	}
	if _, ok := srv.(*HandlerRPCServer); !ok {
		t.Errorf("expected *HandlerRPCServer")
	}

	cli, err := p.Client(&hplugin.MuxBroker{}, &rpc.Client{})
	if err != nil {
		t.Fatalf("Client() failed: %v", err)
	}
	if _, ok := cli.(*HandlerRPCClient); !ok {
		t.Errorf("expected *HandlerRPCClient")
	}
}
