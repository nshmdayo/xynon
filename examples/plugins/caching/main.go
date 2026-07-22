//go:build tinygo.wasm

package main

import (
	"container/list"
	"os"
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

// ---- Cache Logic -----------------------------------------------------------
type cacheEntry struct {
	key        string
	body       string
	statusCode string
	expiresAt  int64
}

var (
	cacheList = list.New()
	cacheMap  = make(map[string]*list.Element)
	maxSize   = 100
	defTTL    = int64(60)
)

func init() {
	if ms := os.Getenv("max_size"); ms != "" {
		if v, err := strconv.Atoi(ms); err == nil && v > 0 {
			maxSize = v
		}
	}
	if dt := os.Getenv("default_ttl"); dt != "" {
		if v, err := strconv.Atoi(dt); err == nil && v > 0 {
			defTTL = int64(v)
		}
	}
}

// ---- Plugin exports --------------------------------------------------------

//export xynon_abi_version
func xynon_abi_version() int32 { return 1 }

//export on_request
func on_request(handle int32) int32 {
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
