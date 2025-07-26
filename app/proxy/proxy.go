package proxy

import (
	"io"
	"log"
	"net/http"
)

func Start(addr string) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Target URL
		targetURL := "http://httpbin:80"

		log.Println("addr: ", addr)
		log.Println("Request from client: ", r)

		// Create new request
		req, err := http.NewRequest(r.Method, targetURL, r.Body)
		if err != nil {
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
