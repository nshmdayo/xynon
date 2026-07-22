package plugin

import (
	"context"
	"net/http"
	"time"

	"github.com/nshmdayo/xynon/internal/plugin/abi"
	"github.com/tetratelabs/wazero/api"
)

// Limits configures sandbox constraints for a plugin.
type Limits struct {
	Timeout  time.Duration // 0 means no limit
	MemoryMB uint32        // 0 means no limit (wazero default)
}

// Plugin is a loaded, ready-to-call WASM plugin instance.
type WasmHandler struct {
	name   string
	mod    api.Module
	limits Limits
	config []byte
}

// Name returns the plugin's declared name.
func (p *WasmHandler) Name() string { return p.name }

// Close releases the plugin's WASM module.
func (p *WasmHandler) Close(ctx context.Context) error {
	return p.mod.Close(ctx)
}

// OnRequest calls the plugin's on_request hook with the given request headers.
// Returns the action and whether a short-circuit was requested (with status code).
func (p *WasmHandler) OnRequest(ctx context.Context, header http.Header) (abi.Action, bool, int, error) {
	if header == nil {
		header = make(http.Header)
	}
	hd := &HandleData{Kind: HandleRequest, Header: header, PluginConfig: p.config}
	return p.callHook(ctx, abi.ExportOnRequest, hd)
}

// OnResponse calls the plugin's on_response hook with the given response headers.
func (p *WasmHandler) OnResponse(ctx context.Context, header http.Header, statusCode int) (abi.Action, bool, int, error) {
	if header == nil {
		header = make(http.Header)
	}
	hd := &HandleData{Kind: HandleResponse, Header: header, StatusCode: statusCode, PluginConfig: p.config}
	return p.callHook(ctx, abi.ExportOnResponse, hd)
}

func (p *WasmHandler) callHook(ctx context.Context, export string, hd *HandleData) (abi.Action, bool, int, error) {
	if p.limits.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.limits.Timeout)
		defer cancel()
	}

	id := allocHandle(hd)
	defer freeHandle(id)

	fn := p.mod.ExportedFunction(export)
	res, err := fn.Call(ctx, uint64(id))
	if err != nil {
		return abi.ActionError, false, 0, &ErrExecution{Cause: err}
	}

	action := abi.Action(int32(res[0]))
	return action, hd.ShortCircuit, hd.ShortCircuitSC, nil
}
