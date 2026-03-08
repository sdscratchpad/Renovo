// fault-injector exposes an HTTP API to programmatically trigger MVP fault scenarios.
// Scenarios: bad-rollout, resource-saturation, batch-timeout.
// Used by demo scripts and the web console to start a demonstration.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	injector "github.com/ravi-poc/fault-injector/internal"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	eventStoreURL := os.Getenv("EVENT_STORE_URL")
	if eventStoreURL == "" {
		eventStoreURL = "http://localhost:8085"
	}

	inj, err := injector.New(eventStoreURL)
	if err != nil {
		log.Fatalf("fault-injector: init: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /scenarios", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(inj.Scenarios())
	})

	mux.HandleFunc("POST /inject/{scenario}", func(w http.ResponseWriter, r *http.Request) {
		scenario := r.PathValue("scenario")
		result, err := inj.Inject(r.Context(), scenario)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(result)
	})

	mux.HandleFunc("POST /restore/{scenario}", func(w http.ResponseWriter, r *http.Request) {
		scenario := r.PathValue("scenario")
		if err := inj.Restore(r.Context(), scenario); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"restored": scenario})
	})

	mux.HandleFunc("POST /restore", func(w http.ResponseWriter, r *http.Request) {
		inj.RestoreAll(r.Context())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "all scenarios restored"})
	})

	log.Printf("fault-injector listening on :%s", port)
	if err := http.ListenAndServe(":"+port, withCORS(mux)); err != nil {
		log.Fatalf("fault-injector exited: %v", err)
	}
}

func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}
