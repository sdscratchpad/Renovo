---
applyTo: "ui/**"
---

# Agent: UI

## Scope
You own `ui/`. Do not touch services or infra.

## Stack
- React 18 + TypeScript. Bootstrapped with Create React App or Vite.
- Plain CSS modules (no CSS-in-JS, no Tailwind).
- No state management library. Use React hooks + fetch for data.

## Pages and components
1. **Dashboard** (`/`) — service health map, active incidents list, live KPI panel.
2. **Incident detail** (`/incidents/:id`) — timeline, AI RCA panel (summary, confidence score, evidence), remediation proposal with approve button.
3. **Fault injector** (`/inject`) — buttons to trigger each of the 3 MVP scenarios.
4. **Audit log** (`/audit`) — chronological audit trail for all incidents.

## API integration
- Stage 1 (agent-ui): build all pages with static mock data in `ui/src/api/mock.ts`.
- Stage 2 (agent-ui-wiring): replace mocks with live calls to real services:
  - `GET http://localhost:8085/incidents` — incident list
  - `GET http://localhost:8083/incidents/{id}/rca` — RCA detail
  - `GET http://localhost:8085/remediations` — remediation list
  - `POST http://localhost:8082/inject/{scenario}` — trigger fault
  - `PATCH http://localhost:8084/remediations/{id}/approve` — approve fix

## KPI panel
- Show: MTTD (minutes), MTTR (minutes), incidents resolved today, auto-resolve rate (%).
- Refresh every 10 seconds.

## Conventions
- All API calls live in `ui/src/api/`. Export typed async functions, never inline fetch calls in components.
- TypeScript types must mirror `contracts/types.go` field names exactly (use camelCase).
- No console.log in production code.
