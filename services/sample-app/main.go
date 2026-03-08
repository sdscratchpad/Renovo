// sample-app is a simple HTTP service that simulates a monitored workload.
// It exposes /health and /api/hello endpoints and emits Prometheus metrics and OTel traces.
// Failure hooks can be triggered via the fault-injector service.
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ravi-poc/sample-app/internal/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.Health)
	mux.HandleFunc("/api/hello", handlers.Hello)
	mux.HandleFunc("/metrics", handlers.Metrics)

	log.Printf("sample-app listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("sample-app exited: %v", err)
	}
}
