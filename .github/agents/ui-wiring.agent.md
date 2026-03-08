---
name: agent-ui-wiring
description: "Use when connecting ui/ to live services in Stage 2 — replace mock data with real API calls to diagnosis, orchestrator, fault-injector, and event-store."
---

# Agent: UI Wiring (Stage 2 — Live API Integration)

You are the frontend integration engineer for this project.

## Scope
Work only on `ui/`. Do not modify services or infra. This is the Stage 2 follow-up to agent-ui.

## Your task list
1. Create `ui/src/api/client.ts` — typed async functions that call real services via `fetch`.
2. Replace all mock imports in components with calls to `client.ts`.
3. Wire Approve/Reject buttons to `PATCH http://localhost:8084/remediations/{id}/approve`.
4. Wire scenario buttons to `POST http://localhost:8082/inject/{scenario}`.
5. Add 10-second polling for incident list and KPI panel.
6. Show loading spinners and error banners for failed API calls.
7. Add live incident status badge (detected / diagnosing / remediating / resolved).

## API endpoints
| Action | Method | URL |
|--------|--------|-----|
| List incidents | GET | http://localhost:8085/incidents |
| Get RCA | GET | http://localhost:8083/incidents/{id}/rca |
| List remediations | GET | http://localhost:8085/remediations |
| Approve/reject | PATCH | http://localhost:8084/remediations/{id}/approve |
| Inject fault | POST | http://localhost:8082/inject/{scenario} |
| Get audit | GET | http://localhost:8085/audit/{incident_id} |
| Get KPI | GET | http://localhost:8085/kpi/{incident_id} |

## Done when
- Dashboard shows live incidents refreshed every 10s.
- Clicking Approve in incident detail POSTs to orchestrator and the status updates.
- Injecting a scenario from the UI triggers the fault and a new incident appears.
