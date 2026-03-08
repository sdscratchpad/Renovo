---
name: agent-ui
description: "Stage 1 kickoff: build ui/ with static mock data — React + TypeScript demo console with all 4 pages."
---

You are the UI agent for Stage 1 of the AI Infrastructure Resilience Copilot build.

Read `.github/instructions/ui.instructions.md` for your full task list and done criteria.

First, bootstrap the React + TypeScript project in `ui/` if it has not been initialized yet:
```
cd ui && npx create-react-app . --template typescript
```

Then build in this order:
1. `ui/src/api/types.ts` — TypeScript types mirroring `contracts/types.go`.
2. `ui/src/api/mock.ts` — static mock data (2 incidents, 1 RCA, 1 remediation, audit entries, KPI).
3. Navigation bar component in `ui/src/components/NavBar.tsx`.
4. Dashboard page `ui/src/pages/Dashboard.tsx` — health cards, incidents list, KPI panel.
5. Incident detail page `ui/src/pages/IncidentDetail.tsx` — RCA panel, evidence list, Approve button.
6. Fault injector page `ui/src/pages/FaultInjector.tsx` — 3 scenario buttons.
7. Audit log page `ui/src/pages/AuditLog.tsx` — audit entries table.
8. Wire routing in `ui/src/App.tsx`.

Use only CSS modules for styling. No Tailwind. When done, run `npm start` and verify all 4 pages render.
