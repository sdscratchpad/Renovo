---
name: agent-orchestrator
description: "Use when building or fixing services/orchestrator/ — Kubernetes runbook executor with policy checks and approval gate."
---

# Agent: Orchestrator

You are the remediation engineer for this project.

## Scope
Work only on `services/orchestrator/`. Do not modify other services.

## Your task list
1. `internal/runbooks/rollback.go` — kubectl rollout undo for the target deployment.
2. `internal/runbooks/restart.go` — delete pods by label selector.
3. `internal/runbooks/scale.go` — scale deployment to specified replica count.
4. `internal/runbooks/retry_batch.go` — POST to batch-worker reset endpoint.
5. `internal/policy.go` — allow-list check (namespace must be `demo`), risk-level approval enforcement.
6. `POST /remediate` — validate policy, execute matching runbook, post result to event-store.
7. `GET /remediations` — list from event-store.
8. `PATCH /remediations/{id}/approve` — set approval status, then auto-execute if approved.

## Policy rules
- Only namespace `demo` is allowed. Return 403 for any other namespace.
- `high` risk: require `approval=approved` or return 403.
- `low`/`medium` risk: auto-execute if `AUTO_MODE=true` env var is set.

## Kubernetes client
- Use `k8s.io/client-go`. Try in-cluster config first, then fall back to `~/.kube/config`.

## Done when
- POST /remediate with `runbook_name: restart-pod` and valid params logs execution and returns result JSON.
- Attempting to target namespace `prod` returns HTTP 403.
