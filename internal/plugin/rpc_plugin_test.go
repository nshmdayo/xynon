package plugin

import (
	"context"
	"net"
	"net/http"
	"net/rpc"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/nshmdayo/xynon/internal/plugin/abi"
)

type mockHandler struct {
	ReqAction abi.Action
	ReqHeader http.Header
	ReqStatus int

	ResAction abi.Action
	ResHeader http.Header
	ResStatus int
}

func (m *mockHandler) Name() string {
	return "mock"
}

func (m *mockHandler) OnRequest(ctx context.Context, header http.Header) (abi.Action, bool, int, error) {
	if m.ReqHeader != nil {
		for k, v := range m.ReqHeader {
			header[k] = v
		}
	}
	return m.ReqAction, m.ReqAction == abi.ActionShortCircuit, m.ReqStatus, nil
}

func (m *mockHandler) OnResponse(ctx context.Context, header http.Header, statusCode int) (abi.Action, bool, int, error) {
	if m.ResHeader != nil {
		for k, v := range m.ResHeader {
			header[k] = v
		}
	}
	return m.ResAction, m.ResAction == abi.ActionShortCircuit, m.ResStatus, nil
}

func (m *mockHandler) Close(ctx context.Context) error {
	return nil
}

func TestRPCClientServer(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	server := rpc.NewServer()
	mock := &mockHandler{
		ReqAction: abi.ActionContinue,
		ReqHeader: http.Header{"X-Mock-Req": []string{"1"}},
		ReqStatus: 0,
		ResAction: abi.ActionShortCircuit,
		ResHeader: http.Header{"X-Mock-Res": []string{"2"}},
		ResStatus: 403,
	}
	err := server.RegisterName("Plugin", &HandlerRPCServer{Impl: mock})
	if err != nil {
		t.Fatalf("Failed to register: %v", err)
	}

	go server.ServeConn(serverConn)

	client := rpc.NewClient(clientConn)
	defer client.Close()

	rpcClient := &HandlerRPCClient{client: client, name: "test"}

	if rpcClient.Name() != "test" {
		t.Errorf("Expected name 'test', got %s", rpcClient.Name())
	}

	// Test OnRequest
	reqHdr := http.Header{"Original": []string{"val"}}
	action, sc, scStatus, err := rpcClient.OnRequest(context.Background(), reqHdr)
	if err != nil {
		t.Fatalf("OnRequest failed: %v", err)
	}
	if action != abi.ActionContinue || sc != false || scStatus != 0 {
		t.Errorf("OnRequest return values incorrect: %v %v %v", action, sc, scStatus)
	}
	if reqHdr.Get("X-Mock-Req") != "1" {
		t.Errorf("Header not modified correctly")
	}

	// Test OnResponse
	resHdr := http.Header{"Original": []string{"val2"}}
	action, sc, scStatus, err = rpcClient.OnResponse(context.Background(), resHdr, 200)
	if err != nil {
		t.Fatalf("OnResponse failed: %v", err)
	}
	if action != abi.ActionShortCircuit || sc != true || scStatus != 403 {
		t.Errorf("OnResponse return values incorrect: %v %v %v", action, sc, scStatus)
	}
	if resHdr.Get("X-Mock-Res") != "2" {
		t.Errorf("Header not modified correctly")
	}
}

func TestXynonPlugin(t *testing.T) {
	p := &XynonPlugin{Impl: &mockHandler{}}

	// Test Server
	srv, err := p.Server(nil)
	if err != nil {
		t.Fatalf("Server() failed: %v", err)
	}
	if _, ok := srv.(*HandlerRPCServer); !ok {
		t.Errorf("Expected *HandlerRPCServer")
	}

	// Test Client
	cli, err := p.Client(nil, nil)
	if err != nil {
		t.Fatalf("Client() failed: %v", err)
	}
	if _, ok := cli.(*HandlerRPCClient); !ok {
		t.Errorf("Expected *HandlerRPCClient")
	}
}

func TestLoadRpcPlugin(t *testing.T) {
	// Build the dummy plugin
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "dummyplugin")
	cmd := exec.Command("go", "build", "-o", binPath, "./testdata/dummyplugin/main.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build dummy plugin: %v", err)
	}

	ctx := context.Background()
	handler, err := LoadRpcPlugin(ctx, "dummy", binPath)
	if err != nil {
		t.Fatalf("LoadRpcPlugin failed: %v", err)
	}
	defer handler.Close(ctx)

	if handler.Name() != "dummy" {
		t.Errorf("Expected name 'dummy', got %v", handler.Name())
	}

	hdr := make(http.Header)
	action, _, _, err := handler.OnRequest(ctx, hdr)
	if err != nil {
		t.Fatalf("OnRequest failed: %v", err)
	}
	if action != abi.ActionContinue {
		t.Errorf("Unexpected action: %v", action)
	}
	if hdr.Get("X-Dummy-Req") != "ok" {
		t.Errorf("Header not set correctly by dummy plugin")
	}

	// Test OnResponse
	action, _, _, err = handler.OnResponse(ctx, hdr, 200)
	if err != nil {
		t.Fatalf("OnResponse failed: %v", err)
	}
	if hdr.Get("X-Dummy-Res") != "ok" {
		t.Errorf("Header not set correctly by dummy plugin")
	}
}
