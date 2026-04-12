//go:build tinygo.wasm

// Package main is a minimal Xynon proxy plugin (ABI v1) built with TinyGo.
// It adds an "X-Xynon: true" header to every proxied request.
package main

import "unsafe"

// ---- ABI constants (must match internal/plugin/abi/abi.go) ----------------

const (
	actionContinue = int32(0)
)

// ---- Host function imports -------------------------------------------------

//go:wasmimport env set_header
func set_header(handle, namePtr, nameLen, valPtr, valLen int32) int32

// ---- Helper ----------------------------------------------------------------

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
	setHeader(handle, "X-Xynon", "true")
	return actionContinue
}

//export on_response
func on_response(_ int32) int32 {
	return actionContinue
}

func main() {}
