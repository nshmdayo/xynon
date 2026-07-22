package plugin

import (
	"context"
	"fmt"
	"time"

	"github.com/nshmdayo/xynon/internal/config"
)

// LoadChain loads WASM and RPC plugins in the given order and returns
// the ordered slice, or an error if any plugin fails to load.
func LoadChain(ctx context.Context, rt *Runtime, entries []ChainEntry, limits Limits) ([]Handler, error) {
	plugins := make([]Handler, 0, len(entries))
	for _, e := range entries {
		var p Handler
		var err error
		if e.Type == config.PluginTypeRPC {
			p, err = LoadRpcPlugin(ctx, e.Name, e.Path)
		} else if e.Type == config.PluginTypeWasm {
			p, err = rt.LoadWasmHandler(ctx, e.Name, e.Path, limits, e.Config)
		} else {
			err = fmt.Errorf("unknown plugin type %q", e.Type)
		}
		if err != nil {
			// Close already loaded plugins before returning.
			for _, already := range plugins {
				already.Close(ctx)
			}
			return nil, fmt.Errorf("load chain: plugin %q: %w", e.Name, err)
		}
		plugins = append(plugins, p)
	}
	return plugins, nil
}

// ChainEntry pairs a plugin name with its execution type and file path.
type ChainEntry struct {
	Name   string
	Type   string
	Path   string
	Config []byte
}

// DefaultLimits returns sensible defaults.
func DefaultLimits() Limits {
	return Limits{
		Timeout:  5 * time.Second,
		MemoryMB: 64,
	}
}
