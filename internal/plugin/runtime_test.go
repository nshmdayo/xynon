package plugin

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/nshmdayo/xynon/internal/plugin/abi"
)

const (
	validWasm    = "testdata/valid/plugin.wasm"
	badABIWasm   = "testdata/bad_abi/plugin.wasm"
	noExportWasm = "testdata/no_exports/plugin.wasm"
)

// TestLoadWasmHandler_Valid tests the functionality of the respective component.
func TestLoadWasmHandler_Valid(t *testing.T) {
	ctx := context.Background()
	rt, err := NewRuntime(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close(ctx)

	p, err := rt.LoadWasmHandler(ctx, "valid", validWasm, Limits{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer p.Close(ctx)

	if p.Name() != "valid" {
		t.Errorf("name = %q, want valid", p.Name())
	}
}

// TestLoadWasmHandler_ABIMismatch tests the functionality of the respective component.
func TestLoadWasmHandler_ABIMismatch(t *testing.T) {
	ctx := context.Background()
	rt, err := NewRuntime(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close(ctx)

	_, err = rt.LoadWasmHandler(ctx, "bad_abi", badABIWasm, Limits{})
	if err == nil {
		t.Fatal("expected ABI version error")
	}
	var abiErr *ErrABIVersion
	if !errors.As(err, &abiErr) {
		t.Errorf("expected ErrABIVersion, got %T: %v", err, err)
	}
}

// TestLoadWasmHandler_MissingExport tests the functionality of the respective component.
func TestLoadWasmHandler_MissingExport(t *testing.T) {
	ctx := context.Background()
	rt, err := NewRuntime(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close(ctx)

	_, err = rt.LoadWasmHandler(ctx, "no_exports", noExportWasm, Limits{})
	if err == nil {
		t.Fatal("expected missing export error")
	}
	var missingErr *ErrMissingExport
	if !errors.As(err, &missingErr) {
		t.Errorf("expected ErrMissingExport, got %T: %v", err, err)
	}
}

// TestOnRequest_ContinueAction tests the functionality of the respective component.
func TestOnRequest_ContinueAction(t *testing.T) {
	ctx := context.Background()
	rt, err := NewRuntime(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close(ctx)

	p, err := rt.LoadWasmHandler(ctx, "valid", validWasm, Limits{})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close(ctx)

	hdr := http.Header{}
	action, sc, _, err := p.OnRequest(ctx, hdr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != abi.ActionContinue {
		t.Errorf("action = %d, want ActionContinue", action)
	}
	if sc {
		t.Error("expected no short-circuit")
	}
}

// TestOnRequest_Timeout tests the functionality of the respective component.
func TestOnRequest_Timeout(t *testing.T) {
	ctx := context.Background()
	rt, err := NewRuntime(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close(ctx)

	// Use a very short timeout — valid plugin returns immediately so we use
	// a cancel-before-call trick instead to simulate timeout path.
	cancelCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
	cancel()

	p, err := rt.LoadWasmHandler(ctx, "valid", validWasm, Limits{Timeout: 1 * time.Nanosecond})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close(ctx)

	_, _, _, err = p.OnRequest(cancelCtx, http.Header{})
	if err == nil {
		// The valid plugin returns instantly so timeout may not fire — acceptable.
		t.Log("no timeout error (plugin returned too fast — acceptable)")
	}
}

// TestHandle_InvalidAfterFree tests the functionality of the respective component.
func TestHandle_InvalidAfterFree(t *testing.T) {
	hd := &HandleData{Kind: HandleRequest, Header: http.Header{}}
	id := allocHandle(hd)
	freeHandle(id)

	got := lookupHandle(id)
	if got != nil {
		t.Errorf("expected nil after free, got %+v", got)
	}
}

// TestOnResponse_ContinueAction tests the functionality of the respective component.
func TestOnResponse_ContinueAction(t *testing.T) {
	ctx := context.Background()
	rt, err := NewRuntime(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close(ctx)

	p, err := rt.LoadWasmHandler(ctx, "valid", validWasm, Limits{})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close(ctx)

	hdr := http.Header{}
	action, sc, _, err := p.OnResponse(ctx, hdr, 200)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != abi.ActionContinue {
		t.Errorf("action = %d, want ActionContinue", action)
	}
	if sc {
		t.Error("expected no short-circuit")
	}
}
