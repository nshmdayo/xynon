package plugin

import (
	"context"
	"fmt"
	"time"
)

// LoadChain loads WASM plugins in the given name-to-path order and returns
// the ordered slice, or an error if any plugin fails to load.
func LoadChain(ctx context.Context, rt *Runtime, entries []ChainEntry, limits Limits) ([]*Plugin, error) {
	plugins := make([]*Plugin, 0, len(entries))
	for _, e := range entries {
		p, err := rt.LoadPlugin(ctx, e.Name, e.Path, limits)
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

// ChainEntry pairs a plugin name with its WASM file path.
type ChainEntry struct {
	Name string
	Path string
}

// DefaultLimits returns sensible defaults.
func DefaultLimits() Limits {
	return Limits{
		Timeout:  5 * time.Second,
		MemoryMB: 64,
	}
}
