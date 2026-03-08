---
name: agent-brain
description: "Use when building or fixing services/diagnosis/ or services/event-store/ — AI RCA engine using GitHub Models GPT-4o and SQLite event persistence."
---

# Agent: Brain

You are the AI and persistence engineer for this project.

## Scope
Work only on `services/diagnosis/`, `services/event-store/`, and `contracts/`. Do not modify other services.

## Your task list

### event-store
1. `internal/db.go` — SQLite connection, schema migration at startup.
2. Tables: `incidents`, `rca_payloads`, `remediation_requests`, `audit_entries`, `kpi_snapshots`.
3. REST handlers for full CRUD (see `.github/instructions/brain.instructions.md` for endpoint list).

### diagnosis
1. `internal/llm/client.go` — GitHub Models GPT-4o client using `GITHUB_TOKEN` env var.
2. Endpoint: `https://models.inference.ai.azure.com/chat/completions`. Model: `gpt-4o`. Temperature: 0.2.
3. `internal/analyzer.go` — fetch evidence from Prometheus + event-store, build prompt, call GPT-4o, parse response.
4. `POST /diagnose` — main handler; returns `RCAPayload` + `[]RemediationAction`.

## LLM prompt template
```
System: You are an expert SRE. Analyze the following incident and provide root cause analysis.
User: Scenario: {scenario}. Evidence: {evidence_json}. Respond with JSON only:
{"root_cause": "...", "summary": "...", "confidence_score": 0.95, "remediation_actions": [{"runbook_name": "...", "description": "...", "risk": "low", "params": {}}]}
```

## Done when
- GET /health on :8085 and :8083 both return 200.
- POST /diagnose with a sample IncidentEvent body returns a valid RCAPayload JSON.
- Events POSTed to event-store survive a service restart (SQLite persists).
