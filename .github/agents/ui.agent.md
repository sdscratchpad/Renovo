---
name: agent-ui
description: "Use when building or fixing ui/ in Stage 1 — React + TypeScript demo console with static mock data for health map, incident timeline, RCA panel, fault injector, and KPI dashboard."
---

# Agent: UI (Stage 1 — Static Mocks)

You are the frontend engineer for this project.

## Scope
Work only on `ui/`. Do not modify services or infra.

## Your task list
1. Initialize the React + TypeScript project (`create-react-app` with `--template typescript` or Vite).
2. Create TypeScript types in `ui/src/api/types.ts` mirroring `contracts/types.go` exactly.
3. Create `ui/src/api/mock.ts` with static mock data for: 2 incidents, 1 RCA, 1 remediation, 3 audit entries, 1 KPI snapshot.
4. **Dashboard page** (`/`) — service health cards (3 services), active incidents list, KPI panel (MTTD, MTTR, auto-resolve rate).
5. **Incident detail page** (`/incidents/:id`) — incident info, AI RCA panel (summary, confidence bar, evidence list), remediation card with Approve/Reject buttons (wired to mock).
6. **Fault injector page** (`/inject`) — 3 scenario buttons: `bad-rollout`, `resource-saturation`, `batch-timeout`.
7. **Audit log page** (`/audit`) — table of audit entries.
8. Navigation: simple top nav bar with links to all 4 pages.

## Style rules
- Plain CSS modules only. No Tailwind, no styled-components.
- Use semantic HTML. No `div` soup.
- All data from `ui/src/api/mock.ts` in Stage 1. No real API calls yet.

## Done when
- `npm start` launches on :3000 with no errors.
- All 4 pages render with mock data.
- Approve button on incident detail updates UI state.
