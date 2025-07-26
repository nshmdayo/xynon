package proxy

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// MITMProxy is a structure that holds the configuration for MITM proxy server
type MITMProxy struct {
	CA      *x509.Certificate
	CAKey   *rsa.PrivateKey
	CertDir string
	Addr    string
	Handler func(*http.Request, *http.Response) // Handler for request/response modification
}

// NewMITMProxy creates a new MITM proxy
func NewMITMProxy(addr string) (*MITMProxy, error) {
	ca, caKey, err := generateCA()
	if err != nil {
		return nil, fmt.Errorf("failed to generate CA: %v", err)
	}

	return &MITMProxy{
		CA:      ca,
		CAKey:   caKey,
		CertDir: "./certs",
		Addr:    addr,
	}, nil
}

// Start starts the MITM proxy server
func (m *MITMProxy) Start() error {
	// Create certificate directory
	if err := os.MkdirAll(m.CertDir, 0755); err != nil {
		return fmt.Errorf("failed to create cert directory: %v", err)
	}

	// CA証明書をファイルに保存
	if err := m.saveCA(); err != nil {
		return fmt.Errorf("failed to save CA: %v", err)
	}

	server := &http.Server{
		Addr:    m.Addr,
		Handler: http.HandlerFunc(m.handleRequest),
	}

	log.Printf("MITM Proxy server starting on %s", m.Addr)
	log.Printf("CA certificate saved to %s/ca.crt", m.CertDir)
	log.Println("Install the CA certificate in your browser to avoid SSL warnings")

	return server.ListenAndServe()
}

// handleRequest は HTTP/HTTPS リクエストを処理する
func (m *MITMProxy) handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == "CONNECT" {
		m.handleConnect(w, r)
	} else {
		m.handleHTTP(w, r)
	}
}

// handleConnect は HTTPS CONNECT メソッドを処理する
func (m *MITMProxy) handleConnect(w http.ResponseWriter, r *http.Request) {
	log.Printf("CONNECT request to %s", r.Host)

	// クライアントに接続確立を通知
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer clientConn.Close()

	// ターゲットサーバーへの接続を確立
	targetConn, err := net.Dial("tcp", r.Host)
	if err != nil {
		log.Printf("Failed to connect to target %s: %v", r.Host, err)
		return
	}
	defer targetConn.Close()

	// サーバー証明書を生成
	cert, err := m.generateCert(r.Host)
	if err != nil {
		log.Printf("Failed to generate certificate for %s: %v", r.Host, err)
		return
	}

	// クライアント側のTLS接続を確立
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*cert},
	}
	clientTLSConn := tls.Server(clientConn, tlsConfig)
	defer clientTLSConn.Close()

	// サーバー側のTLS接続を確立
	serverTLSConn := tls.Client(targetConn, &tls.Config{
		ServerName:         extractHostname(r.Host),
		InsecureSkipVerify: true,
	})
	defer serverTLSConn.Close()

	// TLS ハンドシェイクを実行
	if err := clientTLSConn.Handshake(); err != nil {
		log.Printf("Client TLS handshake failed: %v", err)
		return
	}

	if err := serverTLSConn.Handshake(); err != nil {
		log.Printf("Server TLS handshake failed: %v", err)
		return
	}

	// HTTPS トラフィックを傍受・転送
	m.interceptHTTPS(clientTLSConn, serverTLSConn)
}

// handleHTTP は HTTP リクエストを処理する
func (m *MITMProxy) handleHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("HTTP request to %s", r.URL.String())

	// リクエストを改ざんする機会を提供
	if m.Handler != nil {
		m.Handler(r, nil)
	}

	// ターゲットサーバーにリクエストを転送
	targetURL := r.URL.String()
	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		targetURL = "http://" + r.Host + r.RequestURI
	}

	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// ヘッダーをコピー
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to forward request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// レスポンスを改ざんする機会を提供
	if m.Handler != nil {
		m.Handler(r, resp)
	}

	// レスポンスヘッダーをコピー
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// interceptHTTPS は HTTPS トラフィックを傍受する
func (m *MITMProxy) interceptHTTPS(clientConn, serverConn *tls.Conn) {
	// クライアントからサーバーへの転送（リクエスト）
	go func() {
		defer clientConn.Close()
		defer serverConn.Close()

		reader := bufio.NewReader(clientConn)
		for {
			req, err := http.ReadRequest(reader)
			if err != nil {
				if err != io.EOF {
					log.Printf("Error reading HTTPS request: %v", err)
				}
				break
			}

			log.Printf("HTTPS request: %s %s", req.Method, req.URL.Path)

			// リクエストを改ざんする機会を提供
			if m.Handler != nil {
				m.Handler(req, nil)
			}

			// サーバーにリクエストを転送
			if err := req.Write(serverConn); err != nil {
				log.Printf("Error writing HTTPS request: %v", err)
				break
			}
		}
	}()

	// サーバーからクライアントへの転送（レスポンス）
	reader := bufio.NewReader(serverConn)
	for {
		resp, err := http.ReadResponse(reader, nil)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading HTTPS response: %v", err)
			}
			break
		}

		log.Printf("HTTPS response: %d", resp.StatusCode)

		// レスポンスを改ざんする機会を提供
		if m.Handler != nil {
			m.Handler(nil, resp)
		}

		// クライアントにレスポンスを転送
		if err := resp.Write(clientConn); err != nil {
			log.Printf("Error writing HTTPS response: %v", err)
			break
		}
	}
}

// generateCA は CA証明書と秘密鍵を生成する
func generateCA() (*x509.Certificate, *rsa.PrivateKey, error) {
	// RSA秘密鍵を生成
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// CA証明書テンプレートを作成
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"MITM Proxy"},
			Country:       []string{"JP"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1年間有効
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// 自己署名証明書を作成
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}

	// 証明書をパース
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, err
	}

	return cert, key, nil
}

// generateCert は指定されたホスト名用のサーバー証明書を生成する
func (m *MITMProxy) generateCert(host string) (*tls.Certificate, error) {
	// RSA秘密鍵を生成
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	hostname := extractHostname(host)

	// サーバー証明書テンプレートを作成
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Organization:  []string{"MITM Proxy"},
			Country:       []string{"JP"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    hostname,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{hostname},
	}

	// IP アドレスの場合は IPAddresses に追加
	if ip := net.ParseIP(hostname); ip != nil {
		template.IPAddresses = []net.IP{ip}
	}

	// CA で署名された証明書を作成
	certDER, err := x509.CreateCertificate(rand.Reader, &template, m.CA, &key.PublicKey, m.CAKey)
	if err != nil {
		return nil, err
	}

	// TLS証明書を作成
	cert := tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  key,
	}

	return &cert, nil
}

// saveCA は CA証明書をファイルに保存する
func (m *MITMProxy) saveCA() error {
	// CA証明書をPEM形式で保存
	certFile, err := os.Create(fmt.Sprintf("%s/ca.crt", m.CertDir))
	if err != nil {
		return err
	}
	defer certFile.Close()

	return pem.Encode(certFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: m.CA.Raw,
	})
}

// extractHostname はホスト:ポート形式からホスト名を抽出する
func extractHostname(host string) string {
	hostname, _, err := net.SplitHostPort(host)
	if err != nil {
		return host
	}
	return hostname
}

// SetHandler はリクエスト・レスポンス改ざん用のハンドラーを設定する
func (m *MITMProxy) SetHandler(handler func(*http.Request, *http.Response)) {
	m.Handler = handler
}
