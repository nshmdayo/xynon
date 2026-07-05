package plugin

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/nshmdayo/xynon/internal/plugin/abi"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// Runtime owns the wazero runtime and compilation cache shared across all
// plugin instances.
type Runtime struct {
	rt    wazero.Runtime
	cache wazero.CompilationCache
}

// NewRuntime initialises the wazero runtime with a process-scoped compilation
// cache and WASI support pre-instantiated.
func NewRuntime(ctx context.Context) (*Runtime, error) {
	cache := wazero.NewCompilationCache()
	cfg := wazero.NewRuntimeConfig().WithCompilationCache(cache)
	rt := wazero.NewRuntimeWithConfig(ctx, cfg)

	// Instantiate WASI so plugins compiled with TinyGo (wasip1) can run.
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, rt); err != nil {
		_ = rt.Close(ctx)
		return nil, fmt.Errorf("runtime: instantiate WASI: %w", err)
	}

	// Register host functions.
	if _, err := rt.NewHostModuleBuilder("env").
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(hostGetHeader), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		Export(abi.FnGetHeader).
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(hostSetHeader), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		Export(abi.FnSetHeader).
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(hostSetStatus), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		Export(abi.FnSetStatus).
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(hostShortCircuit), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		Export(abi.FnShortCircuit).
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(hostLog), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		Export(abi.FnLog).
		Instantiate(ctx); err != nil {
		_ = rt.Close(ctx)
		return nil, fmt.Errorf("runtime: instantiate env: %w", err)
	}

	return &Runtime{rt: rt, cache: cache}, nil
}

// Close releases all wazero resources.
func (r *Runtime) Close(ctx context.Context) {
	_ = r.rt.Close(ctx)
	_ = r.cache.Close(ctx)
}

// LoadWasmHandler compiles and instantiates the WASM file at path, registers host
// functions, validates exports, and returns a ready WasmHandler.
func (r *Runtime) LoadWasmHandler(ctx context.Context, name, path string, limits Limits) (*WasmHandler, error) {
	wasmBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("plugin %q: read file: %w", name, err)
	}



	compiled, err := r.rt.CompileModule(ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("plugin %q: compile: %w", name, err)
	}

	// WithStartFunctions() with no args prevents wazero from auto-calling
	// _start (WASI command mode), which would run main() and exit the module.
	// Plugins are WASI reactors: they export hooks and must stay alive.
	modCfg := wazero.NewModuleConfig().
		WithName(name).
		WithStartFunctions(). // do not call _start
		WithStdout(os.Stdout).
		WithStderr(os.Stderr)
	mod, err := r.rt.InstantiateModule(ctx, compiled, modCfg)
	if err != nil {
		return nil, fmt.Errorf("plugin %q: instantiate: %w", name, err)
	}

	// Validate ABI version.
	abiVerFn := mod.ExportedFunction(abi.ExportABIVersion)
	if abiVerFn == nil {
		_ = mod.Close(ctx)
		return nil, &ErrMissingExport{Name: abi.ExportABIVersion}
	}
	res, err := abiVerFn.Call(ctx)
	if err != nil {
		_ = mod.Close(ctx)
		return nil, fmt.Errorf("plugin %q: call %s: %w", name, abi.ExportABIVersion, err)
	}
	got := int32(res[0])
	if got != abi.CurrentVersion {
		_ = mod.Close(ctx)
		return nil, &ErrABIVersion{Got: got, Want: abi.CurrentVersion}
	}

	// Validate required exports.
	for _, fn := range []string{abi.ExportOnRequest, abi.ExportOnResponse} {
		if mod.ExportedFunction(fn) == nil {
			_ = mod.Close(ctx)
			return nil, &ErrMissingExport{Name: fn}
		}
	}

	slog.Info("plugin loaded", "name", name)
	return &WasmHandler{name: name, mod: mod, limits: limits}, nil
}

// ---- host functions -------------------------------------------------------
// Each function receives params/results via the stack slice (in-place).

func hostGetHeader(_ context.Context, mod api.Module, stack []uint64) {
	handle, namePtr, nameLen, outPtr, outCap :=
		int32(stack[0]), uint32(stack[1]), uint32(stack[2]), uint32(stack[3]), uint32(stack[4])

	hd := lookupHandle(handle)
	if hd == nil {
		stack[0] = 0xFFFFFFFF // -1 as uint32
		return
	}

	name := readString(mod.Memory(), namePtr, nameLen)
	val := hd.Header.Get(name)
	if val == "" {
		stack[0] = 0
		return
	}

	n := uint32(len(val))
	if n > outCap {
		n = outCap
	}
	mod.Memory().Write(outPtr, []byte(val)[:n])
	stack[0] = uint64(n)
}

func hostSetHeader(_ context.Context, mod api.Module, stack []uint64) {
	handle, namePtr, nameLen, valPtr, valLen :=
		int32(stack[0]), uint32(stack[1]), uint32(stack[2]), uint32(stack[3]), uint32(stack[4])

	hd := lookupHandle(handle)
	if hd == nil {
		stack[0] = 0xFFFFFFFF
		return
	}

	name := readString(mod.Memory(), namePtr, nameLen)
	val := readString(mod.Memory(), valPtr, valLen)
	hd.Header.Set(name, val)
	stack[0] = 0
}

func hostSetStatus(_ context.Context, _ api.Module, stack []uint64) {
	handle, status := int32(stack[0]), int(stack[1])
	hd := lookupHandle(handle)
	if hd == nil {
		stack[0] = 0xFFFFFFFF
		return
	}
	hd.StatusCode = status
	stack[0] = 0
}

func hostShortCircuit(_ context.Context, _ api.Module, stack []uint64) {
	handle, status := int32(stack[0]), int(stack[1])
	hd := lookupHandle(handle)
	if hd == nil {
		stack[0] = 0xFFFFFFFF
		return
	}
	hd.ShortCircuit = true
	hd.ShortCircuitSC = status
	stack[0] = 0
}

func hostLog(_ context.Context, mod api.Module, stack []uint64) {
	_, msgPtr, msgLen := int32(stack[0]), uint32(stack[1]), uint32(stack[2])
	msg := readString(mod.Memory(), msgPtr, msgLen)
	slog.Info("[plugin]", "msg", msg)
	stack[0] = 0
}

func readString(mem api.Memory, ptr, length uint32) string {
	if length == 0 {
		return ""
	}
	b, ok := mem.Read(ptr, length)
	if !ok {
		return ""
	}
	return string(b)
}
