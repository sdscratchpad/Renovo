---
name: agent-setup
description: "Stage 0 kickoff: scaffold monorepo, contracts, and Copilot customization files."
---

You are running Stage 0 of the AI Infrastructure Resilience Copilot build.

Your job is to validate and complete the Stage 0 scaffold. The following should already exist — verify each one and create any that are missing:

1. `contracts/types.go` — shared Go types: IncidentEvent, RCAPayload, RemediationAction, AuditEntry, KPISnapshot.
2. `contracts/go.mod` — module `github.com/ravi-poc/contracts`.
3. `go.mod` stubs for all 5 services (sample-app, batch-worker, fault-injector, diagnosis, orchestrator, event-store) — each with a `replace` directive pointing to `../../contracts`.
4. `main.go` stubs for all 5 services — minimal HTTP server setup with `/health` handler.
5. Root `Makefile` with `up`, `down`, `reset`, `build`, `test` targets.
6. Per-service `Makefile` with `build`, `up`, `down`, `test` targets.
7. `.github/copilot-instructions.md` — project overview, service ports, stack conventions, GitHub Models auth pattern.
8. `.github/instructions/*.instructions.md` — one per domain (infra, workloads, brain, orchestrator, fault-injector, ui).
9. `.github/agents/*.agent.md` — one per parallel agent.

After verifying, report any gaps found and the current state of the scaffold.
