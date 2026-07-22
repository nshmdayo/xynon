package plugin

import (
	"net/http"
	"sync"
	"sync/atomic"
)

// HandleKind distinguishes request handles from response handles.
type HandleKind uint8

const (
	HandleRequest  HandleKind = iota
	HandleResponse HandleKind = iota
)

// HandleData holds the mutable request or response headers for a single hook
// invocation. It is valid only while the hook is executing.
type HandleData struct {
	Kind           HandleKind
	Header         http.Header
	StatusCode     int  // only used for response handles
	ShortCircuit   bool // set by short_circuit host function
	ShortCircuitSC int
	PluginConfig   []byte
}

var (
	handleCounter atomic.Int32
	handleMu      sync.RWMutex
	handles       = map[int32]*HandleData{}
)

// allocHandle registers hd and returns its ID.
func allocHandle(hd *HandleData) int32 {
	id := handleCounter.Add(1)
	handleMu.Lock()
	handles[id] = hd
	handleMu.Unlock()
	return id
}

// freeHandle removes the handle from the table.
func freeHandle(id int32) {
	handleMu.Lock()
	delete(handles, id)
	handleMu.Unlock()
}

// lookupHandle returns the HandleData for id, or nil if not found.
func lookupHandle(id int32) *HandleData {
	handleMu.RLock()
	hd := handles[id]
	handleMu.RUnlock()
	return hd
}
