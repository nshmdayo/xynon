//go:build tinygo.wasm

package main

import (
	"strconv"
	"encoding/json"
	"sync"
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

//go:wasmimport env set_header
func set_header(handle, namePtr, nameLen, valPtr, valLen int32) int32

//go:wasmimport env short_circuit
func short_circuit(handle, status int32) int32

//go:wasmimport env get_plugin_config
func get_plugin_config(handle, outPtr, outCap int32) int32

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
	
	actualLen := get_header(handle, np, nl, outPtr, outCap)
	if actualLen <= 0 {
		return ""
	}
	
	if actualLen > outCap {
		actualLen = outCap
	}
	
	return string(outBuf[:actualLen])
}

func setHeader(handle int32, name, value string) {
	np, nl := strPtr(name)
	vp, vl := strPtr(value)
	set_header(handle, np, nl, vp, vl)
}

func getPluginConfig(handle int32) []byte {
	outBuf := make([]byte, 1024)
	outPtr := int32(uintptr(unsafe.Pointer(&outBuf[0])))
	outCap := int32(len(outBuf))

	actualLen := get_plugin_config(handle, outPtr, outCap)
	if actualLen <= 0 {
		return nil
	}

	if actualLen > outCap {
		actualLen = outCap
	}

	return outBuf[:actualLen]
}

// ---- Plugin Logic --------------------------------------------------------

type Config struct {
	KeyHeader  string `json:"key_header"`
	Capacity   int    `json:"capacity"`
	RefillRate int    `json:"refill_rate"`
}

type Bucket struct {
	Tokens     int
	LastRefill time.Time
}

var (
	mu      sync.Mutex
	buckets = make(map[string]*Bucket)
)

//export xynon_abi_version
func xynon_abi_version() int32 { return 1 }

//export on_request
func on_request(handle int32) int32 {
	cfgData := getPluginConfig(handle)
	cfg := Config{
		KeyHeader:  "X-Forwarded-For",
		Capacity:   10,
		RefillRate: 1,
	}
	if len(cfgData) > 0 {
		_ = json.Unmarshal(cfgData, &cfg)
	}

	key := getHeader(handle, cfg.KeyHeader)
	if key == "" {
		key = "default"
	}

	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	b, ok := buckets[key]
	if !ok {
		b = &Bucket{
			Tokens:     cfg.Capacity,
			LastRefill: now,
		}
		buckets[key] = b
	}

	// Refill
	elapsed := now.Sub(b.LastRefill).Seconds()
	if elapsed > 0 {
		newTokens := int(elapsed * float64(cfg.RefillRate))
		if newTokens > 0 {
			b.Tokens += newTokens
			if b.Tokens > cfg.Capacity {
				b.Tokens = cfg.Capacity
			}
			// Only advance time for the tokens we actually added
			b.LastRefill = b.LastRefill.Add(time.Duration(newTokens) * time.Second / time.Duration(cfg.RefillRate))
		}
	}

	if b.Tokens > 0 {
		b.Tokens--
		return actionContinue
	}

	// Rate limit exceeded
	short_circuit(handle, 429)
	retryAfter := 1
	if cfg.RefillRate > 0 {
		retryAfter = 1
	}
	setHeader(handle, "Retry-After", strconv.Itoa(retryAfter))
	return actionShortCircuit
}

//export on_response
func on_response(handle int32) int32 {
	return actionContinue
}

func main() {}
