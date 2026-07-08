// Package abi defines the stable contract between the Xynon proxy host
// and WASM plugins.
//
// # ABI Version
//
// The current ABI version is 1. A plugin MUST export the function:
//
//	xynon_abi_version() -> i32
//
// and return 1. Plugins declaring an unsupported version are rejected at load time.
//
// # Plugin Exports
//
// A plugin MUST export the following functions:
//
//	on_request(handle i32) -> i32
//	on_response(handle i32) -> i32
//
// The i32 return value is an Action (see Action constants below).
//
// # Host Imports (namespace "env")
//
// The host provides the following functions that plugins may import:
//
//	get_header(handle i32, name_ptr i32, name_len i32, out_ptr i32, out_cap i32) -> i32
//	  Reads the header named by name[name_ptr:name_ptr+name_len] from the
//	  request or response identified by handle.
//	  Writes the value into guest memory at out_ptr (up to out_cap bytes).
//	  Returns the number of bytes written, or -1 on error.
//
//	set_header(handle i32, name_ptr i32, name_len i32, val_ptr i32, val_len i32) -> i32
//	  Sets the header named by name[...] to val[...] on the object identified
//	  by handle. Returns 0 on success, -1 on error.
//
//	set_status(handle i32, status i32) -> i32
//	  Sets the HTTP status code for a response handle.
//	  Returns 0 on success, -1 on error.
//
//	short_circuit(handle i32, status i32) -> i32
//	  Instructs the host to respond immediately with the given HTTP status code
//	  instead of forwarding to upstream. Only valid from on_request.
//	  Returns 0 on success, -1 on error.
//
//	xynon_log(handle i32, msg_ptr i32, msg_len i32) -> i32
//	  Emits a log line from the plugin. Returns 0.
//
//	compute_hmac_sha256(msg_ptr i32, msg_len i32, key_ptr i32, key_len i32, out_ptr i32, out_cap i32) -> i32
//	  Computes HMAC-SHA256 of msg[...] using key[...].
//	  Writes the raw binary MAC into guest memory at out_ptr (up to out_cap bytes).
//	  Returns the number of bytes written (32), or -1 on error (e.g. out_cap < 32).
//
// # Handle Lifetime
//
// Each invocation of on_request / on_response receives a unique handle i32.
// The handle is valid only for the duration of that single hook invocation.
// After the hook returns the handle is invalidated; using it in a subsequent
// call will return an error code.
//
// # Action Constants
//
// Plugins return one of the following i32 values from hook functions:
//
//	ActionContinue   (0)  — proceed normally
//	ActionShortCircuit (1) — stop the chain; host uses the short_circuit response
//	ActionError      (2)  — plugin error; host skips this plugin for this request
package abi

const (
	// CurrentVersion is the ABI version this host implements.
	CurrentVersion = 1

	// Host function names in the "env" namespace.
	FnGetHeader       = "get_header"
	FnSetHeader       = "set_header"
	FnSetStatus       = "set_status"
	FnShortCircuit    = "short_circuit"
	FnLog             = "xynon_log"
	FnComputeHMAC256  = "compute_hmac_sha256"

	// Plugin export names.
	ExportABIVersion = "xynon_abi_version"
	ExportOnRequest  = "on_request"
	ExportOnResponse = "on_response"
)

// Action is the i32 value returned by a plugin hook.
type Action int32

const (
	ActionContinue     Action = 0
	ActionShortCircuit Action = 1
	ActionError        Action = 2
)
