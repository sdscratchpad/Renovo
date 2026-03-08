package handlers

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Hello handles GET /api/hello. Increments the request counter and injects
// errors at the rate configured by the FORCE_ERROR_RATE env var (0.0-1.0).
func Hello(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	_, span := tracer.Start(r.Context(), "hello")
	defer span.End()

	requestsTotal.Inc()

	errorRate, _ := strconv.ParseFloat(os.Getenv("FORCE_ERROR_RATE"), 64)
	if errorRate > 0 && rand.Float64() < errorRate {
		errorsTotal.Inc()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "injected error"})
		requestDuration.WithLabelValues("5xx").Observe(time.Since(start).Seconds())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "hello from sample-app",
	})
	requestDuration.WithLabelValues("2xx").Observe(time.Since(start).Seconds())
}
