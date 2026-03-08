package handlers

import (
	"encoding/json"
	"net/http"
)

// Health handles GET /health and returns service status.
func Health(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "health")
	defer span.End()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"version": "1.0",
	})
}
