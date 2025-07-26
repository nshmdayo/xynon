package proxy

import (
	"io"
	"log"
	"net/http"
)

func Start(addr string) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("addr: ", addr)
		log.Println("Request from client: ", r)

		// For proxy requests, use the original request URL
		var targetURL string
		if r.URL.IsAbs() {
			// This is a proper proxy request with absolute URL
			targetURL = r.URL.String()
		} else {
			// This is a direct request, construct the target URL
			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			targetURL = scheme + "://" + r.Host + r.URL.Path
			if r.URL.RawQuery != "" {
				targetURL += "?" + r.URL.RawQuery
			}
		}

		log.Printf("Forwarding request to: %s", targetURL)

		// Create new request
		req, err := http.NewRequest(r.Method, targetURL, r.Body)
		if err != nil {
			log.Printf("Failed to create request: %v", err)
			http.Error(w, "Failed to create request", http.StatusInternalServerError)
			return
		}

		// Copy headers from original request
		for key, values := range r.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
		log.Println("Request to target: ", req)

		// Send request using client
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Failed to forward request: %v", err)
			http.Error(w, "Failed to forward request", http.StatusInternalServerError)
			return
		}
		log.Println("Response from target: ", resp)
		defer resp.Body.Close()

		// Return target response to client
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		log.Println("Response to client: ", w)
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	return http.ListenAndServe(addr, nil)
}
