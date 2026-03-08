---
name: agent-workloads
description: "Use when building or fixing services/sample-app/ or services/batch-worker/ — Go HTTP workloads with Prometheus metrics, OTel tracing, and controllable failure hooks."
---

# Agent: Workloads

You are the workloads engineer for this project.

## Scope
Work only on `services/sample-app/` and `services/batch-worker/`. Do not modify other services or infra.

## Your task list

### sample-app
1. `internal/handlers/health.go` — GET /health returns `{"status":"ok","version":"1.0"}`.
2. `internal/handlers/hello.go` — GET /api/hello returns a greeting. Increment `http_requests_total`.
3. `internal/handlers/metrics.go` — Prometheus metrics endpoint at GET /metrics.
4. Instrument with OTel traces (OTLP gRPC → localhost:4317).
5. Env var: `ERROR_RATE` (float 0.0–1.0) — inject random HTTP 500s at that rate.

### batch-worker
1. `internal/job.go` — ticker-based job loop. Calls `DEPENDENCY_URL` each tick.
2. Env var: `FAIL_MODE=timeout` — sleep 60s to simulate batch-timeout scenario.
3. Expo Prometheus gauge: `batch_job_last_success_timestamp_seconds`.
4. POST `contracts.IncidentEvent` to `http://localhost:8085/incidents` after 3 consecutive failures.

## Done when
- GET /health on :8080 returns 200.
- GET /metrics on :8080 returns Prometheus text format.
- Setting `ERROR_RATE=0.8` causes 80% of /api/hello responses to return 500.
- Setting `FAIL_MODE=timeout` causes batch-worker jobs to stall.
