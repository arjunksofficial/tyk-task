package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func main() {
	backend := os.Getenv("BACKEND_URL")
	if backend == "" {
		log.Fatal("BACKEND_URL env var required")
	}
	target, err := url.Parse(backend)
	if err != nil {
		log.Fatalf("Invalid BACKEND_URL: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	http.HandleFunc("/api/v1/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})
	log.Println("Proxy listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
