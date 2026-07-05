package proxy

import (
	"sync/atomic"

	"github.com/nshmdayo/xynon/internal/plugin"
)

// Chain is an ordered, immutable list of loaded plugins for a single
// generation. It is safe to hold a reference across a hot-reload boundary.
type Chain struct {
	plugins []plugin.Handler
}

// Plugins returns the ordered slice of plugins in this chain.
func (c *Chain) Plugins() []plugin.Handler {
	if c == nil {
		return nil
	}
	return c.plugins
}

// NewChain constructs a Chain from an ordered slice of plugins.
func NewChain(plugins []plugin.Handler) *Chain {
	return &Chain{plugins: plugins}
}

// ChainRegistry holds the currently active Chain behind an atomic pointer so
// that readers (request handlers) never block and hot-reload writers simply
// do a single Store().
type ChainRegistry struct {
	ptr atomic.Pointer[Chain]
}

// Load returns the current Chain. May return nil if no chain has been stored.
func (r *ChainRegistry) Load() *Chain {
	return r.ptr.Load()
}

// Store atomically replaces the active chain. The previous chain remains
// accessible to in-flight requests via their held *Chain reference until
// those requests finish.
func (r *ChainRegistry) Store(c *Chain) {
	r.ptr.Store(c)
}
