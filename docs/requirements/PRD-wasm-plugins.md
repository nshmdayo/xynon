# Product Requirements Document: WASM Plugins (`WasmHandler`)

## Feature Description
Lightweight, in-process plugins executed via the WebAssembly runtime.

## Requirements
- **Runtime**: `wazero` (Zero-dependency WebAssembly runtime).
- **Compilation Target**: `wasip1` (WASI Snapshot Preview 1) via TinyGo.
- **Data Passing**: Communication between the Go host and the WASM guest is done via explicit memory sharing and exported ABI functions.
- **Constraints**: 
  - WASM plugins must be compiled using Go 1.23+ and TinyGo.
  - Avoid imports of packages that use CGO or dependencies incompatible with WebAssembly targets.
