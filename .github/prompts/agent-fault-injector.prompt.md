---
name: agent-fault-injector
description: "Stage 2 kickoff: build services/fault-injector — HTTP API to trigger bad-rollout, resource-saturation, and batch-timeout scenarios on the kind cluster."
---

You are the Fault Injector agent for Stage 2 of the AI Infrastructure Resilience Copilot build.

Read `.github/instructions/fault-injector.instructions.md` for your full task list and done criteria.

Build in this order:
1. `services/fault-injector/internal/scenarios/bad_rollout.go` — patch sample-app Deployment to introduce a bad env var.
2. `services/fault-injector/internal/scenarios/resource_saturation.go` — patch resource limits to near-zero.
3. `services/fault-injector/internal/scenarios/batch_timeout.go` — patch batch-worker to set FAIL_MODE=timeout.
4. `services/fault-injector/internal/injector.go` — dispatcher that routes scenario name to the correct handler.
5. Wire `POST /inject/{scenario}`, `GET /scenarios`, `GET /health` in `services/fault-injector/main.go`.
6. After each injection, POST an `IncidentEvent` to `http://localhost:8085/incidents`.

When done, run `make -C services/fault-injector up` and test: `curl -X POST http://localhost:8082/inject/batch-timeout`.
