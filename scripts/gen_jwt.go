//go:build ignore

package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

func main() {
	secret := []byte("xynon-secret-key")

	header := map[string]interface{}{
		"alg": "HS256",
		"typ": "JWT",
	}
	payload := map[string]interface{}{
		"sub": "test-user",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	}

	hBytes, _ := json.Marshal(header)
	pBytes, _ := json.Marshal(payload)

	hBase64 := base64.RawURLEncoding.EncodeToString(hBytes)
	pBase64 := base64.RawURLEncoding.EncodeToString(pBytes)

	message := hBase64 + "." + pBase64
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	fmt.Printf("%s.%s\n", message, signature)
}
