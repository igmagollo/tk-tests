package main

import (
	"log"
	"net/http"
	"os"

	"github.com/igmagollo/tk-tests/internal/api"
)

func main() {
	addr := ":" + envOr("PORT", "8080")
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, api.New().Routes()); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
