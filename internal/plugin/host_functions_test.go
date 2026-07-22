package plugin

import (
	"context"
	"net/http"
	"testing"

	"github.com/tetratelabs/wazero/api"
)

type mockMemory struct {
	api.Memory // embed to satisfy interface for methods we don't mock
	data       []byte
}

func (m *mockMemory) Read(offset, byteCount uint32) ([]byte, bool) {
	if offset+byteCount > uint32(len(m.data)) {
		return nil, false
	}
	return m.data[offset : offset+byteCount], true
}

func (m *mockMemory) Write(offset uint32, v []byte) bool {
	if offset+uint32(len(v)) > uint32(len(m.data)) {
		return false
	}
	copy(m.data[offset:], v)
	return true
}

type mockModule struct {
	api.Module
	mem *mockMemory
}

func (m *mockModule) Memory() api.Memory {
	return m.mem
}

func TestHostFunctions(t *testing.T) {
	ctx := context.Background()

	// Setup memory
	mem := &mockMemory{data: make([]byte, 1024)}
	mod := &mockModule{mem: mem}

	// Write a test string "X-Test" at offset 10
	copy(mem.data[10:], "X-Test")
	// Write "Value" at offset 20
	copy(mem.data[20:], "Value")

	// Create a handle
	req, _ := http.NewRequest("GET", "http://localhost", nil)
	req.Header.Set("X-Test", "HelloWorld")
	
	hd := &HandleData{
		Kind:         HandleRequest,
		Header:       req.Header,
		PluginConfig: []byte("config-data"),
	}
	id := allocHandle(hd)
	defer freeHandle(id)

	t.Run("hostGetHeader", func(t *testing.T) {
		stack := []uint64{uint64(id), 10, 6, 100, 50} // handle, namePtr, nameLen, outPtr, outCap
		hostGetHeader(ctx, mod, stack)
		if stack[0] != 10 {
			t.Errorf("expected length 10, got %d", stack[0])
		}
		res := string(mem.data[100:110])
		if res != "HelloWorld" {
			t.Errorf("expected HelloWorld, got %s", res)
		}
	})

	t.Run("hostSetHeader", func(t *testing.T) {
		stack := []uint64{uint64(id), 10, 6, 20, 5} // handle, namePtr, nameLen, valPtr, valLen
		hostSetHeader(ctx, mod, stack)
		if hd.Header.Get("X-Test") != "Value" {
			t.Errorf("expected Value, got %s", hd.Header.Get("X-Test"))
		}
	})

	t.Run("hostSetStatus", func(t *testing.T) {
		stack := []uint64{uint64(id), 404}
		hostSetStatus(ctx, mod, stack)
		if hd.StatusCode != 404 {
			t.Errorf("expected status 404, got %d", hd.StatusCode)
		}
	})

	t.Run("hostShortCircuit", func(t *testing.T) {
		stack := []uint64{uint64(id), 403}
		hostShortCircuit(ctx, mod, stack)
		if !hd.ShortCircuit || hd.ShortCircuitSC != 403 {
			t.Errorf("expected short circuit with 403")
		}
	})

	t.Run("hostLog", func(t *testing.T) {
		stack := []uint64{uint64(id), 20, 5}
		hostLog(ctx, mod, stack)
	})

	t.Run("hostComputeHMAC256", func(t *testing.T) {
		copy(mem.data[30:], "msg")
		copy(mem.data[40:], "key")
		stack := []uint64{30, 3, 40, 3, 200, 32} // msgPtr, msgLen, keyPtr, keyLen, outPtr, outCap
		hostComputeHMAC256(ctx, mod, stack)
		if stack[0] != 32 {
			t.Errorf("expected 32, got %d", stack[0])
		}
		// verify some output was written
		empty := true
		for _, b := range mem.data[200:232] {
			if b != 0 {
				empty = false
				break
			}
		}
		if empty {
			t.Errorf("expected non-empty HMAC output")
		}
	})

	t.Run("hostGetPluginConfig", func(t *testing.T) {
		stack := []uint64{uint64(id), 300, 50}
		hostGetPluginConfig(ctx, mod, stack)
		if stack[0] != 11 {
			t.Errorf("expected 11, got %d", stack[0])
		}
		if string(mem.data[300:311]) != "config-data" {
			t.Errorf("expected config-data, got %s", string(mem.data[300:311]))
		}
	})
	
	t.Run("InvalidHandle", func(t *testing.T) {
		stack := []uint64{9999, 10, 6, 100, 50}
		hostGetHeader(ctx, mod, stack)
		if stack[0] != 0xFFFFFFFF {
			t.Errorf("expected 0xFFFFFFFF, got %x", stack[0])
		}
	})
}
