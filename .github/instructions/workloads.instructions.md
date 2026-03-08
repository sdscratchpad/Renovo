---
applyTo: "services/sample-app/**,services/batch-worker/**"
---

# Workloads Agent Instructions

## Scope
You own `services/sample-app/` and `services/batch-worker/`. Do not modify other services.

## What to build

### sample-app
- `internal/handlers/health.go` — GET /health returns `{"status":"ok","version":"1.0"}`.
- `internal/handlers/hello.go` — GET /api/hello returns a greeting and increments a request counter metric.
- `internal/handlers/metrics.go` — Prometheus metrics handler (expose `/metrics`).
- Instrument with OTel traces on every handler.
- Expose Prometheus counters: `sample_app_requests_total`, `sample_app_errors_total`.
- Expose a `FORCE_ERROR_RATE` env var (float 0.0–1.0). When set, inject random errors at that rate.

### batch-worker
- `internal/job.go` — Job logic that reads from a simulated in-memory queue, processes items with a short sleep, and emits OTel spans.
- Expose `BLOCK_DEPENDENCY` env var (bool). When `true`, simulate a blocked dependency by sleeping indefinitely inside the job, causing timeouts.
- Expose Prometheus gauge: `batch_worker_queue_depth`, counter: `batch_worker_jobs_processed_total`.

## Key constraints
- Use only Go standard library + OTel SDK + Prometheus client.
- All errors logged with `log.Printf("sample-app: %v", err)`.
- All handlers return JSON with Content-Type: application/json.
- Import shared types from `github.com/ravi-poc/contracts`.

## Done criteria
- `make -C services/sample-app up` starts the service and GET /health returns 200.
- `make -C services/batch-worker up` starts the worker and logs job processing every 10 seconds.
- `/metrics` on :8080 returns Prometheus metrics.
