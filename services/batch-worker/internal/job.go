// Package internal contains the core job logic for batch-worker.
package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/ravi-poc/contracts"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	queueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "batch_worker_queue_depth",
		Help: "Current depth of the simulated job queue.",
	})
	jobsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "batch_worker_jobs_processed_total",
		Help: "Total number of batch jobs processed successfully.",
	})

	tracer trace.Tracer

	// consecutiveFailures tracks uninterrupted job failures for incident detection.
	consecutiveFailures atomic.Int64
)

const failureThreshold = 3

func init() {
	tracer = otel.Tracer("batch-worker")
}

// isBlocked returns true when the worker should simulate a blocked dependency.
// Triggered by BLOCK_DEPENDENCY=true or FAIL_MODE=timeout.
func isBlocked() bool {
	return os.Getenv("BLOCK_DEPENDENCY") == "true" || os.Getenv("FAIL_MODE") == "timeout"
}

// ProcessJob processes a single item from the simulated queue.
// It records an OTel span and updates Prometheus metrics.
func ProcessJob(ctx context.Context) error {
	_, span := tracer.Start(ctx, "process-job")
	defer span.End()

	queueDepth.Inc()
	defer queueDepth.Dec()

	if isBlocked() {
		log.Printf("batch-worker: dependency blocked")
		<-ctx.Done()
		return fmt.Errorf("batch-worker: dependency timeout: %w", ctx.Err())
	}

	time.Sleep(100 * time.Millisecond)
	jobsProcessed.Inc()
	consecutiveFailures.Store(0)
	log.Printf("batch-worker: job processed successfully")
	return nil
}

// RecordFailure increments the consecutive failure counter and pushes an
// IncidentEvent to the event-store once the failure threshold is crossed.
func RecordFailure(err error) {
	n := consecutiveFailures.Add(1)
	log.Printf("batch-worker: consecutive failures=%d: %v", n, err)
	if n >= failureThreshold {
		pushIncident()
	}
}

func pushIncident() {
	storeURL := os.Getenv("EVENT_STORE_URL")
	if storeURL == "" {
		storeURL = "http://localhost:8085"
	}
	event := contracts.IncidentEvent{
		Scenario:   "batch-timeout",
		Service:    "batch-worker",
		Namespace:  "default",
		Severity:   contracts.SeverityHigh,
		DetectedAt: time.Now().UTC(),
		Evidence: []contracts.EvidenceItem{{
			Source:    "logs",
			Signal:    "consecutive_failures",
			Value:     fmt.Sprintf("%d", consecutiveFailures.Load()),
			Timestamp: time.Now().UTC(),
		}},
	}
	body, err := json.Marshal(event)
	if err != nil {
		log.Printf("batch-worker: failed to marshal incident: %v", err)
		return
	}
	resp, err := http.Post(storeURL+"/incidents", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("batch-worker: failed to push incident: %v", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("batch-worker: incident pushed, status=%d", resp.StatusCode)
}
