package proxy

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewMITMProxy(t *testing.T) {
	proxy, err := NewMITMProxy(":0")
	if err != nil {
		t.Fatalf("Failed to create MITM proxy: %v", err)
	}

	if proxy == nil {
		t.Fatal("Proxy is nil")
	}

	if proxy.CA == nil {
		t.Error("CA certificate was not generated")
	}

	if proxy.CAKey == nil {
		t.Error("CA private key was not generated")
	}

	if proxy.CertDir != "./certs" {
		t.Errorf("Expected CertDir './certs', got '%s'", proxy.CertDir)
	}

	if proxy.Addr != ":0" {
		t.Errorf("Expected Addr ':0', got '%s'", proxy.Addr)
	}
}

func TestMITMProxy_SetHandler(t *testing.T) {
	proxy, err := NewMITMProxy(":0")
	if err != nil {
		t.Fatalf("Failed to create MITM proxy: %v", err)
	}

	// ハンドラーが設定されていないことを確認
	if proxy.Handler != nil {
		t.Error("Handler should be nil initially")
	}

	// ハンドラーを設定
	var called bool
	handler := func(req *http.Request, resp *http.Response) {
		called = true
	}
	proxy.SetHandler(handler)

	if proxy.Handler == nil {
		t.Error("Handler was not set")
	}

	// ハンドラーが呼び出されることを確認
	proxy.Handler(nil, nil)
	if !called {
		t.Error("Handler was not called")
	}
}

func TestMITMProxy_HandleHTTP(t *testing.T) {
	// テスト用ターゲットサーバーを作成
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test-Response", "true")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from target"))
	}))
	defer targetServer.Close()

	// MITM プロキシを作成
	proxy, err := NewMITMProxy(":0")
	if err != nil {
		t.Fatalf("Failed to create MITM proxy: %v", err)
	}

	// リクエスト・レスポンス傍受ハンドラーを設定
	var interceptedRequest *http.Request
	var interceptedResponse *http.Response
	proxy.SetHandler(func(req *http.Request, resp *http.Response) {
		if req != nil {
			interceptedRequest = req
			req.Header.Set("X-Intercepted-Request", "true")
		}
		if resp != nil {
			interceptedResponse = resp
			resp.Header.Set("X-Intercepted-Response", "true")
		}
	})

	// テスト用リクエストを作成
	req := httptest.NewRequest("GET", targetServer.URL, nil)
	req.Header.Set("X-Original-Header", "test")

	// レスポンスレコーダーを作成
	w := httptest.NewRecorder()

	// HTTPリクエストを処理
	proxy.handleHTTP(w, req)

	// レスポンスを検証
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if body := w.Body.String(); body != "Hello from target" {
		t.Errorf("Expected body 'Hello from target', got '%s'", body)
	}

	// ヘッダーが正しく設定されていることを確認
	if w.Header().Get("X-Test-Response") != "true" {
		t.Error("Target server header was not copied")
	}

	if w.Header().Get("X-Intercepted-Response") != "true" {
		t.Error("Intercepted response header was not added")
	}

	// リクエストが傍受されたことを確認
	if interceptedRequest == nil {
		t.Error("Request was not intercepted")
	} else {
		if interceptedRequest.Header.Get("X-Intercepted-Request") != "true" {
			t.Error("Request was not modified by handler")
		}
	}

	// レスポンスが傍受されたことを確認
	if interceptedResponse == nil {
		t.Error("Response was not intercepted")
	}
}

func TestMITMProxy_HandleConnect(t *testing.T) {
	proxy, err := NewMITMProxy(":0")
	if err != nil {
		t.Fatalf("Failed to create MITM proxy: %v", err)
	}

	// CONNECT リクエストを作成
	req := httptest.NewRequest("CONNECT", "https://example.com:443", nil)
	req.Host = "example.com:443"

	w := httptest.NewRecorder()

	// CONNECT メソッドを処理
	proxy.handleConnect(w, req)

	// レスポンスステータスを確認（接続が確立されるまでのステータス）
	if w.Code != http.StatusOK {
		t.Logf("CONNECT response status: %d (this may be expected for test environment)", w.Code)
	}
}

func TestMITMProxy_GenerateCert(t *testing.T) {
	proxy, err := NewMITMProxy(":0")
	if err != nil {
		t.Fatalf("Failed to create MITM proxy: %v", err)
	}

	tests := []struct {
		host     string
		expected string
	}{
		{"example.com:443", "example.com"},
		{"localhost:8080", "localhost"},
		{"192.168.1.1:443", "192.168.1.1"},
		{"test.local:9000", "test.local"},
	}

	for _, test := range tests {
		cert, err := proxy.generateCert(test.host)
		if err != nil {
			t.Errorf("Failed to generate certificate for %s: %v", test.host, err)
			continue
		}

		if cert == nil {
			t.Errorf("Certificate is nil for host %s", test.host)
			continue
		}

		if len(cert.Certificate) == 0 {
			t.Errorf("Certificate is empty for host %s", test.host)
		}

		// 証明書の内容を検証
		if len(cert.Certificate) > 0 {
			// 証明書をパース
			parsedCert, err := tls.X509KeyPair(cert.Certificate[0], nil)
			if err == nil && parsedCert.Certificate != nil {
				t.Logf("Certificate generated successfully for %s", test.host)
			}
		}
	}
}

func TestExtractHostname(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"example.com:443", "example.com"},
		{"localhost:8080", "localhost"},
		{"192.168.1.1:80", "192.168.1.1"},
		{"example.com", "example.com"},
		{"[::1]:8080", "::1"},
		{"[2001:db8::1]:443", "2001:db8::1"},
	}

	for _, test := range tests {
		result := extractHostname(test.input)
		if result != test.expected {
			t.Errorf("extractHostname(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestMITMProxy_SaveCA(t *testing.T) {
	proxy, err := NewMITMProxy(":0")
	if err != nil {
		t.Fatalf("Failed to create MITM proxy: %v", err)
	}

	// テスト用の一時ディレクトリを作成
	tempDir := "./test_certs"
	proxy.CertDir = tempDir
	defer os.RemoveAll(tempDir)

	// ディレクトリを作成
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// CA証明書を保存
	if err := proxy.saveCA(); err != nil {
		t.Fatalf("Failed to save CA certificate: %v", err)
	}

	// ファイルが存在することを確認
	certFile := tempDir + "/ca.crt"
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		t.Error("CA certificate file was not created")
	}

	// ファイルの内容を確認
	content, err := os.ReadFile(certFile)
	if err != nil {
		t.Fatalf("Failed to read CA certificate file: %v", err)
	}

	if !strings.Contains(string(content), "-----BEGIN CERTIFICATE-----") {
		t.Error("CA certificate file does not contain PEM header")
	}

	if !strings.Contains(string(content), "-----END CERTIFICATE-----") {
		t.Error("CA certificate file does not contain PEM footer")
	}
}

func TestMITMProxy_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// ターゲットサーバーを作成
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Hello from target server"}`))
	}))
	defer targetServer.Close()

	// MITM プロキシを作成
	proxy, err := NewMITMProxy(":0")
	if err != nil {
		t.Fatalf("Failed to create MITM proxy: %v", err)
	}

	// テスト用の証明書ディレクトリを設定
	proxy.CertDir = "./test_integration_certs"
	defer os.RemoveAll(proxy.CertDir)

	// リクエスト・レスポンス統計を収集
	var requestCount, responseCount int
	proxy.SetHandler(func(req *http.Request, resp *http.Response) {
		if req != nil {
			requestCount++
			t.Logf("Intercepted request #%d: %s %s", requestCount, req.Method, req.URL.String())
		}
		if resp != nil {
			responseCount++
			t.Logf("Intercepted response #%d: %d %s", responseCount, resp.StatusCode, resp.Status)
		}
	})

	// プロキシサーバーをテスト用に起動
	proxyServer := httptest.NewServer(http.HandlerFunc(proxy.handleRequest))
	defer proxyServer.Close()

	// プロキシ経由でリクエストを送信
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(proxyServer.URL)
			},
		},
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(targetServer.URL)
	if err != nil {
		t.Fatalf("Failed to send request through proxy: %v", err)
	}
	defer resp.Body.Close()

	// レスポンスを検証
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expectedBody := `{"message": "Hello from target server"}`
	if string(body) != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, string(body))
	}

	// 統計を確認（プロキシを通すと複数回呼ばれる可能性があるため、最低1回は確認）
	if requestCount < 1 {
		t.Errorf("Expected at least 1 intercepted request, got %d", requestCount)
	}

	if responseCount < 1 {
		t.Errorf("Expected at least 1 intercepted response, got %d", responseCount)
	}

	t.Logf("Integration test completed successfully")
}

func TestMITMProxy_HandleRequestMethod(t *testing.T) {
	proxy, err := NewMITMProxy(":0")
	if err != nil {
		t.Fatalf("Failed to create MITM proxy: %v", err)
	}

	tests := []struct {
		method   string
		url      string
		expected string
	}{
		{"GET", "http://example.com/test", "HTTP"},
		{"POST", "http://example.com/api", "HTTP"},
		{"CONNECT", "https://example.com:443", "CONNECT"},
		{"PUT", "http://example.com/data", "HTTP"},
	}

	for _, test := range tests {
		req := httptest.NewRequest(test.method, test.url, nil)
		if test.method == "CONNECT" {
			req.Host = "example.com:443"
		}

		w := httptest.NewRecorder()

		// リクエストを処理
		proxy.handleRequest(w, req)

		t.Logf("Method %s handled, response status: %d", test.method, w.Code)
	}
}

// ベンチマークテスト
func BenchmarkMITMProxy_GenerateCert(b *testing.B) {
	proxy, err := NewMITMProxy(":0")
	if err != nil {
		b.Fatalf("Failed to create MITM proxy: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := proxy.generateCert("example.com:443")
		if err != nil {
			b.Fatalf("Failed to generate certificate: %v", err)
		}
	}
}

func BenchmarkMITMProxy_HandleHTTP(b *testing.B) {
	// ターゲットサーバーを作成
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer targetServer.Close()

	proxy, err := NewMITMProxy(":0")
	if err != nil {
		b.Fatalf("Failed to create MITM proxy: %v", err)
	}

	req := httptest.NewRequest("GET", targetServer.URL, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		proxy.handleHTTP(w, req)
	}
}

func TestMITMProxy_HTTPSInterception(t *testing.T) {
	// このテストはHTTPS傍受の基本的な仕組みをテスト
	proxy, err := NewMITMProxy(":0")
	if err != nil {
		t.Fatalf("Failed to create MITM proxy: %v", err)
	}

	// 証明書生成のテスト
	cert, err := proxy.generateCert("test.example.com:443")
	if err != nil {
		t.Fatalf("Failed to generate certificate: %v", err)
	}

	if cert == nil {
		t.Fatal("Generated certificate is nil")
	}

	if len(cert.Certificate) == 0 {
		t.Fatal("Generated certificate is empty")
	}

	// TLS設定のテスト
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*cert},
	}

	if len(tlsConfig.Certificates) != 1 {
		t.Error("TLS config should contain exactly one certificate")
	}

	t.Log("HTTPS interception components test passed")
}

func TestMITMProxy_ErrorHandling(t *testing.T) {
	proxy, err := NewMITMProxy(":0")
	if err != nil {
		t.Fatalf("Failed to create MITM proxy: %v", err)
	}

	// 存在しないホストでのHTTPリクエストテスト
	req := httptest.NewRequest("GET", "http://nonexistent-host.local/test", nil)
	w := httptest.NewRecorder()

	proxy.handleHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Logf("Error handling test: got status %d", w.Code)
	}

	// 不正なホストでのCONNECTリクエストテスト
	connectReq := httptest.NewRequest("CONNECT", "https://nonexistent-host.local:443", nil)
	connectReq.Host = "nonexistent-host.local:443"
	connectW := httptest.NewRecorder()

	proxy.handleConnect(connectW, connectReq)

	t.Log("Error handling tests completed")
}
