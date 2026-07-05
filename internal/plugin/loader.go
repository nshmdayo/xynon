package plugin

import (
	"context"
	"fmt"
	"time"
)

// LoadChain loads WASM and RPC plugins in the given order and returns
// the ordered slice, or an error if any plugin fails to load.
func LoadChain(ctx context.Context, rt *Runtime, entries []ChainEntry, limits Limits) ([]Handler, error) {
	plugins := make([]Handler, 0, len(entries))
	for _, e := range entries {
		var p Handler
		var err error
		if e.Type == "rpc" {
			p, err = LoadRpcPlugin(ctx, e.Name, e.Path)
		} else {
			p, err = rt.LoadPlugin(ctx, e.Name, e.Path, limits)
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
	Name string
	Type string
	Path string
}

// DefaultLimits returns sensible defaults.
func DefaultLimits() Limits {
	return Limits{
		Timeout:  5 * time.Second,
		MemoryMB: 64,
	}
}
