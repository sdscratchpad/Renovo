package handlers

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics serves the Prometheus metrics endpoint at /metrics.
func Metrics(w http.ResponseWriter, r *http.Request) {
	promhttp.Handler().ServeHTTP(w, r)
}
