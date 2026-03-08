// batch-worker is a recurring batch job that processes items from a simulated queue.
// It exposes controllable failure hooks (e.g. dependency timeout) for fault injection scenarios.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	worker "github.com/ravi-poc/batch-worker/internal"
)

func main() {
	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "9091"
	}

	// Expose Prometheus metrics on a separate port.
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		log.Printf("batch-worker: metrics on :%s", metricsPort)
		if err := http.ListenAndServe(":"+metricsPort, mux); err != nil {
			log.Printf("batch-worker: metrics server error: %v", err)
		}
	}()

	log.Println("batch-worker started")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		case <-ticker.C:
			// Job context with a 30s timeout to detect hung dependencies.
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			err := worker.ProcessJob(ctx)
			cancel()
			if err != nil {
				log.Printf("batch-worker: job failed: %v", err)
				worker.RecordFailure(err)
			}
		case <-stop:
			log.Println("batch-worker stopping")
			return
		}
	}
}
