package main

import (
	"flag"
	"log"
	"net/http"
	"regexp"
	"strings"

	"nproxy/app/mock"
	"nproxy/app/proxy"
)

func main() {
	var (
		addr    = flag.String("addr", ":8080", "proxy server address")
		mitm    = flag.Bool("mitm", false, "start as MITM proxy")
		modify  = flag.Bool("modify", false, "enable request/response modification")
		verbose = flag.Bool("v", false, "output detailed logs")
		mockSrv = flag.Bool("mock", false, "start as mock server")
	)
	flag.Parse()

	if *mockSrv {
		// Start mock server
		log.Printf("Starting mock server on %s", *addr)
		if err := mock.Start(*addr); err != nil {
			log.Fatalf("Failed to start mock server: %v", err)
		}
	} else if *mitm {
		// Start MITM proxy
		mitmProxy, err := proxy.NewMITMProxy(*addr)
		if err != nil {
			log.Fatalf("Failed to create MITM proxy: %v", err)
		}

		if *modify {
			// Set request/response modification handler
			mitmProxy.SetHandler(createModificationHandler(*verbose))
		} else if *verbose {
			// Set logging-only handler
			mitmProxy.SetHandler(createLoggingHandler())
		}

		log.Printf("Starting MITM proxy server on %s", *addr)
		if err := mitmProxy.Start(); err != nil {
			log.Fatalf("Failed to start MITM proxy: %v", err)
		}
	} else {
		// Start normal proxy
		log.Printf("Starting simple proxy server on %s", *addr)
		if err := proxy.Start(*addr); err != nil {
			log.Fatalf("Failed to start proxy: %v", err)
		}
	}
}

// createModificationHandler creates a handler for request/response modification
func createModificationHandler(verbose bool) func(*http.Request, *http.Response) {
	return func(req *http.Request, resp *http.Response) {
		if req != nil {
			if verbose {
				log.Printf("Request: %s %s", req.Method, req.URL.String())
				log.Printf("Request Headers: %v", req.Header)
			}

			// Example of request header modification
			req.Header.Set("X-MITM-Proxy", "true")
			req.Header.Set("User-Agent", "MITM-Proxy/1.0")

			// Modify requests for specific patterns
			if strings.Contains(req.URL.Path, "/api/") {
				req.Header.Set("X-API-Modified", "true")
			}
		}

		if resp != nil {
			if verbose {
				log.Printf("Response: %d %s", resp.StatusCode, resp.Status)
				log.Printf("Response Headers: %v", resp.Header)
			}

			// Example of response header modification
			resp.Header.Set("X-MITM-Intercepted", "true")
			resp.Header.Set("X-Proxy-Time", "2024-01-01")

			// Add security headers
			resp.Header.Set("X-Content-Type-Options", "nosniff")
			resp.Header.Set("X-Frame-Options", "DENY")
			resp.Header.Set("X-XSS-Protection", "1; mode=block")

			// Process text/html content type
			if contentType := resp.Header.Get("Content-Type"); strings.Contains(contentType, "text/html") {
				resp.Header.Set("X-HTML-Modified", "true")
			}
		}
	}
}

// createLoggingHandler creates a handler for logging only
func createLoggingHandler() func(*http.Request, *http.Response) {
	return func(req *http.Request, resp *http.Response) {
		if req != nil {
			log.Printf("📤 Request: %s %s", req.Method, req.URL.String())

			// Log headers while masking sensitive information
			logHeaders(req.Header, "Request")
		}

		if resp != nil {
			log.Printf("📥 Response: %d %s", resp.StatusCode, resp.Status)

			// Log response headers
			logHeaders(resp.Header, "Response")
		}
	}
}

// logHeaders safely logs header information
func logHeaders(headers http.Header, prefix string) {
	sensitiveHeaders := []string{
		"Authorization", "Cookie", "Set-Cookie", "X-API-Key", "X-Auth-Token",
	}

	for key, values := range headers {
		// Check if header contains sensitive information
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
