//go:build tinygo.wasm

package main

import (
	"container/list"
	"encoding/json"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

// ---- ABI constants ---------------------------------------------------------
const (
	actionContinue     = int32(0)
	actionShortCircuit = int32(1)
)

// ---- Host function imports -------------------------------------------------
//go:wasmimport env set_header
func set_header(handle, namePtr, nameLen, valPtr, valLen int32) int32

//go:wasmimport env get_header
func get_header(handle, namePtr, nameLen, outPtr, outCap int32) int32

//go:wasmimport env short_circuit
func short_circuit(handle, status int32) int32

//go:wasmimport env xynon_log
func xynon_log(handle, msgPtr, msgLen int32) int32

//go:wasmimport env get_plugin_config
func get_plugin_config(handle, outPtr, outCap int32) int32

// ---- Helpers ---------------------------------------------------------------
func strPtr(s string) (int32, int32) {
	if len(s) == 0 {
		return 0, 0
	}
	return int32(uintptr(unsafe.Pointer(unsafe.SliceData([]byte(s))))), int32(len(s))
}

func bytePtr(b []byte) (int32, int32) {
	if len(b) == 0 {
		return 0, 0
	}
	return int32(uintptr(unsafe.Pointer(&b[0]))), int32(cap(b))
}

func getHeader(handle int32, name string) string {
	np, nl := strPtr(name)
	buf := make([]byte, 4096)
	op, ocap := bytePtr(buf)
	n := get_header(handle, np, nl, op, ocap)
	if n <= 0 {
		return ""
	}
	return string(buf[:n])
}

func setHeader(handle int32, name, value string) {
	np, nl := strPtr(name)
	vp, vl := strPtr(value)
	set_header(handle, np, nl, vp, vl)
}

func logMsg(msg string) {
	p, l := strPtr(msg)
	xynon_log(0, p, l)
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

// ---- Cache Logic -----------------------------------------------------------
type cacheEntry struct {
	key        string
	body       string
	statusCode string
	expiresAt  int64
}

type Config struct {
	MaxSize    string `json:"max_size"`
	DefaultTTL string `json:"default_ttl"`
}

var (
	cacheList   = list.New()
	cacheMap    = make(map[string]*list.Element)
	maxSize     = 100
	defTTL      = int64(60)
	cfgLoaded   = false
)

func loadConfig(handle int32) {
	if cfgLoaded {
		return
	}
	cfgLoaded = true

	cfgData := getPluginConfig(handle)
	if len(cfgData) == 0 {
		return
	}
	var cfg Config
	if err := json.Unmarshal(cfgData, &cfg); err != nil {
		return
	}
	if cfg.MaxSize != "" {
		if v, err := strconv.Atoi(cfg.MaxSize); err == nil && v > 0 {
			maxSize = v
		}
	}
	if cfg.DefaultTTL != "" {
		if v, err := strconv.Atoi(cfg.DefaultTTL); err == nil && v > 0 {
			defTTL = int64(v)
		}
	}
}

// ---- Plugin exports --------------------------------------------------------

//export xynon_abi_version
func xynon_abi_version() int32 { return 1 }

//export on_request
func on_request(handle int32) int32 {
	loadConfig(handle)

	method := getHeader(handle, "X-Xynon-Req-Method")
	cc := getHeader(handle, "Cache-Control")
	uri := getHeader(handle, "X-Xynon-Req-Uri")

	if method != "GET" && method != "" {
		return actionContinue
	}
	if strings.Contains(cc, "no-cache") || strings.Contains(cc, "no-store") {
		return actionContinue
	}
	if uri == "" {
		return actionContinue
	}

	if el, ok := cacheMap[uri]; ok {
		logMsg("on_request cache hit for uri=" + uri)
		entry := el.Value.(*cacheEntry)
		setHeader(handle, "X-Xynon-Res-Body", entry.body)
		short_circuit(handle, int32(200))
		return actionShortCircuit
	} else {
		logMsg("on_request cache miss for uri=" + uri)
	}

	return actionContinue
}

//export on_response
func on_response(handle int32) int32 {
	loadConfig(handle)

	method := getHeader(handle, "X-Xynon-Req-Method")
	cc := getHeader(handle, "Cache-Control")
	uri := getHeader(handle, "X-Xynon-Req-Uri")
	body := getHeader(handle, "X-Xynon-Res-Body")
	status := getHeader(handle, "X-Xynon-Res-Status")

	
	if method != "GET" && method != "" {
		return actionContinue
	}
	if strings.Contains(cc, "no-cache") || strings.Contains(cc, "no-store") {
		return actionContinue
	}
	if uri == "" {
		return actionContinue
	}
	if body == "" {
		logMsg("on_response body is EMPTY for uri=" + uri)
		return actionContinue
	}
	if status != "200" {
		return actionContinue
	}

	
	ttl := defTTL
	if strings.Contains(cc, "max-age=") {
		parts := strings.Split(cc, "max-age=")
		if len(parts) > 1 {
			ageStr := strings.Split(parts[1], ",")[0]
			ageStr = strings.TrimSpace(ageStr)
			if age, err := strconv.Atoi(ageStr); err == nil {
				ttl = int64(age)
			}
		}
	}

	entry := &cacheEntry{
		key:        uri,
		body:       body,
		statusCode: status,
		expiresAt:  time.Now().Unix() + ttl,
	}

	if el, ok := cacheMap[uri]; ok {
		el.Value = entry
		cacheList.MoveToFront(el)
	} else {
		el := cacheList.PushFront(entry)
		cacheMap[uri] = el
		if cacheList.Len() > maxSize {
			last := cacheList.Back()
			if last != nil {
				lastEntry := last.Value.(*cacheEntry)
				cacheList.Remove(last)
				delete(cacheMap, lastEntry.key)
			}
		}
	}
	logMsg("on_response cached body of length " + strconv.Itoa(len(body)) + " for uri=" + uri)

	return actionContinue
}

func main() {}
