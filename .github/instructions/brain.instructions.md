---
applyTo: "services/diagnosis/**,services/event-store/**"
---

# Brain Agent Instructions

## Scope
You own `services/diagnosis/` and `services/event-store/`. Do not modify other services.

## What to build

### event-store
- `internal/db.go` — SQLite connection setup and schema migration (incidents, rca, remediations, audit, kpi tables).
- `internal/handlers/` — REST handlers:
  - `POST /incidents` — store IncidentEvent
  - `GET /incidents` — list all incidents (newest first)
  - `GET /incidents/{id}` — get single incident
  - `POST /rca` — store RCAPayload
  - `GET /rca/{incident_id}` — get RCA for incident
  - `POST /remediations` — store RemediationRequest
  - `GET /remediations/{incident_id}` — get remediation for incident
  - `PATCH /remediations/{id}/approve` — approve or reject
  - `POST /audit` — store AuditEntry
  - `GET /audit/{incident_id}` — get audit trail
  - `GET /kpi/{incident_id}` — get KPI snapshot

### diagnosis
- `internal/llm/client.go` — GitHub Models GPT-4o client. Auth via `GITHUB_TOKEN`. Endpoint: `https://models.inference.ai.azure.com/chat/completions`.
- `internal/analyzer.go` — Fetches supporting evidence from event-store and Prometheus, builds an LLM prompt, calls GPT-4o, parses response into RCAPayload + []RemediationAction.
- Handler: `POST /diagnose` — accepts IncidentEvent JSON, runs analysis, returns RCAPayload.
- Handler: `GET /incidents` — proxy to event-store.
- Handler: `GET /incidents/{id}/rca` — proxy to event-store.

## LLM prompt pattern
```
System: You are an SRE AI. Given incident signals, identify the root cause and recommend one remediation action.
User: Incident: <scenario>. Evidence: <evidence items>. Return JSON: {"summary": "...", "root_cause": "...", "confidence": 0.95, "action": {"runbook": "...", "params": {...}, "risk": "low"}}
```

## Key constraints
- GITHUB_TOKEN from env var, never hardcoded.
- event-store URL from EVENT_STORE_URL env var (default: http://localhost:8085).
- Prometheus URL from PROMETHEUS_URL env var (default: http://localhost:9090).
- All types from `github.com/ravi-poc/contracts`.

## Done criteria
- `make -C services/event-store up` starts and GET /health returns 200.
- `make -C services/diagnosis up` starts and GET /health returns 200.
- POST /diagnose with a sample IncidentEvent returns a valid RCAPayload.
