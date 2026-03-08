---
applyTo: "services/fault-injector/**"
---

# Agent: Fault Injector

## Scope
You own `services/fault-injector/`. Do not touch other services or infra.

## Responsibilities
- `POST /inject/{scenario}` — triggers one of the 3 MVP fault scenarios.
- `GET /health` — JSON health check.
- `GET /scenarios` — list available scenario names and descriptions.

## Scenarios
- `bad-rollout` — patches the sample-app Deployment in the `demo` namespace to set an invalid env var, causing CrashLoopBackOff.
- `resource-saturation` — applies a stress container sidecar to sample-app or sets resource limits to very low values.
- `batch-timeout` — sets `FAIL_MODE=timeout` on the batch-worker Deployment via a patch.

## Implementation
- Use `k8s.io/client-go` to apply patches against the local kind cluster.
- Each scenario lives in its own file under `internal/scenarios/`.
- After injecting, POST a `contracts.IncidentEvent` to event-store at `http://localhost:8085/incidents` so the UI shows it immediately.

## Conventions
- Return `{"scenario": "...", "injected_at": "..."}` on success, `{"error": "..."}` on failure.
- Each scenario must be idempotent — calling it again resets the fault first.
