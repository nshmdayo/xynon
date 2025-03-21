package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProxy(t *testing.T) {
	// モックサーバーを作成
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Mock-Header", "mockValue")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "Hello from target server")
	}))
	defer targetServer.Close()

	// プロキシサーバーを作成
	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Scheme = "http"
		r.URL.Host = strings.TrimPrefix(targetServer.URL, "http://")
		Start(r.URL.String())
	}))
	defer proxyServer.Close()

	// プロキシサーバーにリクエストを送信
	resp, err := http.Get(proxyServer.URL)
	if err != nil {
		t.Fatalf("Failed to send request to proxy server: %v", err)
	}
	defer resp.Body.Close()

	// レスポンスを検証
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	if resp.Header.Get("X-Mock-Header") != "mockValue" {
		t.Errorf("Expected X-Mock-Header to be 'mockValue', got '%s'", resp.Header.Get("X-Mock-Header"))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if string(body) != "Hello from target server" {
		t.Errorf("Expected response body 'Hello from target server', got '%s'", string(body))
	}
}
