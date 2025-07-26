package main

import (
	"flag"
	"log"
	"net/http"
	"regexp"
	"strings"

	"nproxy/app/proxy"
)

func main() {
	var (
		addr    = flag.String("addr", ":8080", "プロキシサーバーのアドレス")
		mitm    = flag.Bool("mitm", false, "MITMプロキシとして起動")
		modify  = flag.Bool("modify", false, "リクエスト・レスポンスの改ざんを有効にする")
		verbose = flag.Bool("v", false, "詳細ログを出力")
	)
	flag.Parse()

	if *mitm {
		// MITM プロキシを起動
		mitmProxy, err := proxy.NewMITMProxy(*addr)
		if err != nil {
			log.Fatalf("Failed to create MITM proxy: %v", err)
		}

		if *modify {
			// リクエスト・レスポンス改ざんハンドラーを設定
			mitmProxy.SetHandler(createModificationHandler(*verbose))
		} else if *verbose {
			// ログ出力のみのハンドラーを設定
			mitmProxy.SetHandler(createLoggingHandler())
		}

		log.Printf("Starting MITM proxy server on %s", *addr)
		if err := mitmProxy.Start(); err != nil {
			log.Fatalf("Failed to start MITM proxy: %v", err)
		}
	} else {
		// 通常のプロキシを起動
		log.Printf("Starting simple proxy server on %s", *addr)
		if err := proxy.Start(*addr); err != nil {
			log.Fatalf("Failed to start proxy: %v", err)
		}
	}
}

// createModificationHandler はリクエスト・レスポンス改ざん用のハンドラーを作成
func createModificationHandler(verbose bool) func(*http.Request, *http.Response) {
	return func(req *http.Request, resp *http.Response) {
		if req != nil {
			if verbose {
				log.Printf("Request: %s %s", req.Method, req.URL.String())
				log.Printf("Request Headers: %v", req.Header)
			}

			// リクエストヘッダーの改ざん例
			req.Header.Set("X-MITM-Proxy", "true")
			req.Header.Set("User-Agent", "MITM-Proxy/1.0")

			// 特定のパターンのリクエストを書き換え
			if strings.Contains(req.URL.Path, "/api/") {
				req.Header.Set("X-API-Modified", "true")
			}
		}

		if resp != nil {
			if verbose {
				log.Printf("Response: %d %s", resp.StatusCode, resp.Status)
				log.Printf("Response Headers: %v", resp.Header)
			}

			// レスポンスヘッダーの改ざん例
			resp.Header.Set("X-MITM-Intercepted", "true")
			resp.Header.Set("X-Proxy-Time", "2024-01-01")

			// セキュリティヘッダーの追加
			resp.Header.Set("X-Content-Type-Options", "nosniff")
			resp.Header.Set("X-Frame-Options", "DENY")
			resp.Header.Set("X-XSS-Protection", "1; mode=block")

			// Content-Typeが text/html の場合の処理
			if contentType := resp.Header.Get("Content-Type"); strings.Contains(contentType, "text/html") {
				resp.Header.Set("X-HTML-Modified", "true")
			}
		}
	}
}

// createLoggingHandler はログ出力専用のハンドラーを作成
func createLoggingHandler() func(*http.Request, *http.Response) {
	return func(req *http.Request, resp *http.Response) {
		if req != nil {
			log.Printf("📤 Request: %s %s", req.Method, req.URL.String())

			// 機密情報をマスクしてヘッダーをログ出力
			logHeaders(req.Header, "Request")
		}

		if resp != nil {
			log.Printf("📥 Response: %d %s", resp.StatusCode, resp.Status)

			// レスポンスヘッダーをログ出力
			logHeaders(resp.Header, "Response")
		}
	}
}

// logHeaders はヘッダー情報を安全にログ出力する
func logHeaders(headers http.Header, prefix string) {
	sensitiveHeaders := []string{
		"Authorization", "Cookie", "Set-Cookie", "X-API-Key", "X-Auth-Token",
	}

	for key, values := range headers {
		// 機密情報を含むヘッダーかチェック
		isSensitive := false
		for _, sensitive := range sensitiveHeaders {
			if matched, _ := regexp.MatchString("(?i)"+sensitive, key); matched {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			log.Printf("  %s Header %s: [MASKED]", prefix, key)
		} else {
			log.Printf("  %s Header %s: %v", prefix, key, values)
		}
	}
}
