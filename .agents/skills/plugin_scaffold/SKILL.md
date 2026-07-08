---
name: plugin_scaffold
description: Scaffolds boilerplate code for Xynon WASM or RPC plugins.
---

# Plugin Scaffold Skill

When the user asks you to create a new plugin for Xynon, use this skill to generate the correct boilerplate.

## 1. WASM Plugin Scaffold
To create a WASM plugin, create a file under `examples/plugins/<plugin_name>/main.go` using `write_to_file`.
Boilerplate:
```go
//go:build tinygo.wasm

package main

import "unsafe"

// ---- ABI constants ---------------------------------------------------------
const (
	actionContinue = int32(0)
)

// ---- Host function imports -------------------------------------------------
//go:wasmimport env set_header
func set_header(handle, namePtr, nameLen, valPtr, valLen int32) int32

// ---- Helpers ---------------------------------------------------------------
func strPtr(s string) (int32, int32) {
	if len(s) == 0 {
		return 0, 0
	}
	return int32(uintptr(unsafe.Pointer(unsafe.SliceData([]byte(s))))), int32(len(s))
}

func setHeader(handle int32, name, value string) {
	np, nl := strPtr(name)
	vp, vl := strPtr(value)
	set_header(handle, np, nl, vp, vl)
}

// ---- Plugin exports --------------------------------------------------------

//export xynon_abi_version
func xynon_abi_version() int32 { return 1 }

//export on_request
func on_request(handle int32) int32 {
	// e.g. setHeader(handle, "X-My-Header", "value")
	return actionContinue
}

//export on_response
func on_response(handle int32) int32 {
	return actionContinue
}

func main() {}
```

## 2. RPC Plugin Scaffold
To create an RPC plugin, create a file under `examples/plugins/<plugin_name>/main.go`:
```go
package main

import (
	"context"
	"net/http"

	hplugin "github.com/hashicorp/go-plugin"
	"github.com/nshmdayo/xynon/internal/plugin"
	"github.com/nshmdayo/xynon/internal/plugin/abi"
)

type MyPlugin struct{}

func (m *MyPlugin) Name() string { return "<plugin_name>" }

func (m *MyPlugin) OnRequest(ctx context.Context, header http.Header) (abi.Action, bool, int, error) {
	return abi.ActionContinue, false, 0, nil
}

func (m *MyPlugin) OnResponse(ctx context.Context, header http.Header, statusCode int) (abi.Action, bool, int, error) {
	return abi.ActionContinue, false, 0, nil
}

func (m *MyPlugin) Close(ctx context.Context) error { return nil }

func main() {
	hplugin.Serve(&hplugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig,
		Plugins: map[string]hplugin.Plugin{
			"handler": &plugin.XynonPlugin{Impl: &MyPlugin{}},
		},
	})
}
```

## 3. Configuration Update
After creating a plugin, remember to instruct the user or update `examples/config.yaml` to include the plugin. 
For WASM: `type: "wasm"`
For RPC: `type: "rpc"`
