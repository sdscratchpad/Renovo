// Package handlers contains HTTP handlers and shared instrumentation for sample-app.
package handlers

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Prometheus metrics exposed by this service.
var (
	requestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "sample_app_requests_total",
		Help: "Total number of requests handled by sample-app.",
	})
	errorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "sample_app_errors_total",
		Help: "Total number of error responses from sample-app.",
	})
	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "sample_app_request_duration_seconds",
		Help:    "Duration of HTTP requests to sample-app in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"status"})
)

// tracer is the OTel tracer for this service. Uses the global provider (noop by default in Stage 1).
var tracer trace.Tracer

func init() {
	tracer = otel.Tracer("sample-app")
}
