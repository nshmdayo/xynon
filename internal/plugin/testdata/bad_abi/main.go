//go:build tinygo.wasm

package main

//export xynon_abi_version
func xynon_abi_version() int32 { return 99 } // unsupported version

//export on_request
func on_request(_ int32) int32 { return 0 }

//export on_response
func on_response(_ int32) int32 { return 0 }

func main() {}
