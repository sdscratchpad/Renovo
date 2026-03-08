---
name: agent-workloads
description: "Stage 1 kickoff: build services/sample-app and services/batch-worker — Go HTTP workloads with Prometheus metrics, OTel tracing, and controllable failure hooks."
---

You are the Workloads agent for Stage 1 of the AI Infrastructure Resilience Copilot build.

Read `.github/instructions/workloads.instructions.md` for your full task list and done criteria.

Start with sample-app:
1. `services/sample-app/internal/handlers/health.go`
2. `services/sample-app/internal/handlers/hello.go`
3. `services/sample-app/internal/handlers/metrics.go`
4. Wire all handlers in `services/sample-app/main.go`.
5. Add Prometheus instrumentation and OTel tracing.
6. Implement `ERROR_RATE` env var behaviour.
7. Write unit tests.

Then batch-worker:
1. `services/batch-worker/internal/job.go`
2. Implement `FAIL_MODE=timeout` behaviour.
3. Implement consecutive failure tracking and IncidentEvent push.
4. Wire in `services/batch-worker/main.go`.
5. Write unit tests.

When done, run `make -C services/sample-app up` and verify GET /health returns 200.
