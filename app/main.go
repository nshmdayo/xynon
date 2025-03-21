package main

import (
	"log"
	"nproxy/app/proxy"
)

func main() {
	log.Println("Starting proxy server on :8000")
	if err := proxy.Start(":8000"); err != nil {
		log.Fatalf("Error starting proxy server: %v", err)
	}
}
