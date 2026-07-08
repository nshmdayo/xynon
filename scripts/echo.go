//go:build ignore

package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		for k, vv := range r.Header {
			for _, v := range vv {
				fmt.Fprintf(w, "%s: %s\n", k, v)
			}
		}
	})
	http.ListenAndServe(":8081", nil)
}
