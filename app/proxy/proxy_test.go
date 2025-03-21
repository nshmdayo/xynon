package proxy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProxy(t *testing.T) {
	// モックサーバーを作成
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
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
}
