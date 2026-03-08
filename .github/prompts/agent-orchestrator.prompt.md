---
name: agent-orchestrator
description: "Stage 2 kickoff: build services/orchestrator — Kubernetes runbook executor with policy checks and approval gate."
---

You are the Orchestrator agent for Stage 2 of the AI Infrastructure Resilience Copilot build.

Read `.github/instructions/orchestrator.instructions.md` for your full task list and done criteria.

Build in this order:
1. `services/orchestrator/internal/policy.go` — namespace allow-list check, risk-level approval enforcement.
2. `services/orchestrator/internal/runbooks/rollback.go` — rollout undo via client-go.
3. `services/orchestrator/internal/runbooks/restart.go` — delete pods by label selector.
4. `services/orchestrator/internal/runbooks/scale.go` — scale deployment replicas.
5. `services/orchestrator/internal/runbooks/retry_batch.go` — trigger batch-worker reset.
6. `services/orchestrator/internal/executor.go` — dispatch action to matching runbook, post result to event-store.
7. Wire `POST /remediate`, `GET /remediations`, `PATCH /remediations/{id}/approve` in `main.go`.

Policy reminder: only namespace `demo` is allowed. High-risk actions need explicit approval.

When done, run `make -C services/orchestrator up` and test with a sample RemediationAction JSON.
