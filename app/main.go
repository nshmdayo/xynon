package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	fmt.Println("Starting server on port :8000")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 転送先のURL
		targetURL := "http://httpbin:80"

		// 新しいリクエストを作成
		req, err := http.NewRequest(r.Method, targetURL, r.Body)
		if err != nil {
			http.Error(w, "Failed to create request", http.StatusInternalServerError)
			return
		}

		// 元のリクエストのヘッダーをコピー
		for key, values := range r.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		// クライアントを使ってリクエストを送信
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Failed to forward request", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// 転送先のレスポンスをクライアントに返す
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)

		fmt.Println("Successfully forwarded request " + time.Now().Format(time.RFC3339))
	})

	http.ListenAndServe(":8000", nil)
}
