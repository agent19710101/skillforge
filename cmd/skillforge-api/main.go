package main

import (
	"log"
	"net/http"
	"os"

	"github.com/agent19710101/skillforge/internal/api"
	"github.com/agent19710101/skillforge/internal/catalog"
)

func main() {
	addr := envOrDefault("SKILLFORGE_LISTEN_ADDR", ":8080")
	repoRoot := envOrDefault("SKILLFORGE_REPO_ROOT", ".")

	index, err := catalog.BuildIndex(repoRoot)
	if err != nil {
		log.Fatalf("build catalog index: %v", err)
	}

	handler := api.NewServer(index).Handler()
	log.Printf("skillforge-api listening on %s (repo root: %s)", addr, repoRoot)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
