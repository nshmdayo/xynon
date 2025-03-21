package proxy

import (
	"io"
	"log"
	"net/http"
)

func Start(addr string) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 転送先のURL
		targetURL := "http://httpbin:80"

		log.Println("addr: ", addr)
		log.Println("Request from client: ", r)

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
		log.Println("Request to target: ", req)

		// クライアントを使ってリクエストを送信
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Failed to forward request", http.StatusInternalServerError)
			return
		}
		log.Println("Response from target: ", resp)
		defer resp.Body.Close()

		// 転送先のレスポンスをクライアントに返す
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
