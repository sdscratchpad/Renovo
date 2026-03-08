---
applyTo: "services/orchestrator/**"
---

# Agent: Orchestrator

## Scope
You own `services/orchestrator/`. Do not touch other services or infra.

## Responsibilities
- `POST /remediate` — accepts `contracts.RemediationAction`, validates policy, executes runbook.
- `GET /remediations` — list all remediation requests from event-store.
- `PATCH /remediations/{id}/approve` — updates approval status to "approved" in event-store, then executes.
- After execution, POST a `contracts.RemediationResult` to event-store and POST an `contracts.AuditEntry`.

## Runbooks (implement in `internal/runbooks/`)
- `rollback-deployment` — kubectl rollout undo deployment/<name> -n <namespace>
- `restart-pod` — kubectl delete pod -l app=<name> -n <namespace>
- `scale-deployment` — kubectl scale deployment/<name> --replicas=<n> -n <namespace>
- `retry-batch-job` — POST to batch-worker's internal reset endpoint

## Policy rules
- Only namespaces `demo` is in the allow-list. Reject any action targeting other namespaces.
- `high` risk actions require `approval = "approved"` in the request; return 403 otherwise.
- `low` and `medium` risk actions can auto-execute if `AUTO_MODE=true` env var is set.

## Kubernetes client
- Use `k8s.io/client-go`. Load in-cluster config first; fall back to kubeconfig at `~/.kube/config`.
- Use `exec.Command("kubectl", ...)` only as a fallback for runbooks not supported by client-go.
