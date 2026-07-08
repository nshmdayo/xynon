//go:build tinygo.wasm

package main

import (

	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
	"unsafe"
)

// ---- ABI constants ---------------------------------------------------------
const (
	actionContinue     = int32(0)
	actionShortCircuit = int32(1)
	actionError        = int32(2)
)

// ---- Host function imports -------------------------------------------------
//go:wasmimport env get_header
func get_header(handle, namePtr, nameLen, outPtr, outCap int32) int32

//go:wasmimport env short_circuit
func short_circuit(handle, status int32) int32

//go:wasmimport env compute_hmac_sha256
func compute_hmac_sha256(msgPtr, msgLen, keyPtr, keyLen, outPtr, outCap int32) int32

// ---- Helpers ---------------------------------------------------------------
func strPtr(s string) (int32, int32) {
	if len(s) == 0 {
		return 0, 0
	}
	return int32(uintptr(unsafe.Pointer(unsafe.SliceData([]byte(s))))), int32(len(s))
}

func getHeader(handle int32, name string) string {
	np, nl := strPtr(name)
	outBuf := make([]byte, 256)
	outPtr := int32(uintptr(unsafe.Pointer(&outBuf[0])))
	outCap := int32(len(outBuf))
	
	// get_header returns the actual length of the header value
	actualLen := get_header(handle, np, nl, outPtr, outCap)
	if actualLen <= 0 {
		return ""
	}
	
	// If the actual length is greater than the buffer, we would normally allocate a larger buffer and try again.
	// For simplicity in this example, we cap it at the buffer length.
	if actualLen > outCap {
		actualLen = outCap
	}
	
	return string(outBuf[:actualLen])
}

// ---- Plugin exports --------------------------------------------------------

const secretKey = "xynon-secret-key"

func verifyJWT(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	
	// Verify signature using host API
	message := parts[0] + "." + parts[1]
	mp, ml := strPtr(message)
	kp, kl := strPtr(secretKey)
	
	outBuf := make([]byte, 32)
	outPtr := int32(uintptr(unsafe.Pointer(&outBuf[0])))
	outCap := int32(len(outBuf))
	
	n := compute_hmac_sha256(mp, ml, kp, kl, outPtr, outCap)
	if n != 32 {
		return false
	}
	
	expectedSig := base64.RawURLEncoding.EncodeToString(outBuf)
	
	if parts[2] != expectedSig {
		return false
	}
	
	// Verify expiration
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	
	var payload struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return false
	}
	
	// time.Now() in WASM defaults to 1970 if the host doesn't inject it via WASI.
	// But TinyGo targeting wasip1 will use the host's walltime if provided.
	if payload.Exp > 0 && time.Now().Unix() > payload.Exp {
		return false
	}
	
	return true
}

//export xynon_abi_version
func xynon_abi_version() int32 { return 1 }

//export on_request
func on_request(handle int32) int32 {
	auth := getHeader(handle, "Authorization")
	
	if !strings.HasPrefix(auth, "Bearer ") {
		short_circuit(handle, 401)
		return actionShortCircuit
	}
	
	token := auth[7:]
	if !verifyJWT(token) {
		short_circuit(handle, 401)
		return actionShortCircuit
	}
	
	return actionContinue
}

//export on_response
func on_response(handle int32) int32 {
	return actionContinue
}

func main() {}
