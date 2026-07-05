---
name: xynon_scaffold
description: Scaffolds boilerplate code for Xynon WASM or RPC plugins.
---

# Xynon Scaffold Skill

When the user asks you to create a new plugin for Xynon, use this skill to generate the correct boilerplate.

## 1. WASM Plugin Scaffold
To create a WASM plugin, create a file under `examples/plugins/<plugin_name>/main.go` using `write_to_file`.
Boilerplate:
```go
package main

import (
	"github.com/nshmdayo/xynon/internal/plugin/abi"
)

//export xynon_abi_version
func xynonABIVersion() int32 {
	return abi.CurrentVersion
}

//export on_request
func onRequest(id uint64) int32 {
	// e.g. abi.HostSetHeader(id, "X-My-Header", "value")
	return int32(abi.ActionContinue)
}

//export on_response
func onResponse(id uint64) int32 {
	return int32(abi.ActionContinue)
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
