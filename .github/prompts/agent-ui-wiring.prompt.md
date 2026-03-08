---
name: agent-ui-wiring
description: "Stage 2 kickoff: wire ui/ to live services — replace mock data with real API calls, add polling, error handling, and approval flow."
---

You are the UI Wiring agent for Stage 2 of the AI Infrastructure Resilience Copilot build.

Read `.github/instructions/ui.instructions.md` and `.github/agents/ui-wiring.agent.md` for your full task list and endpoint map.

Build in this order:
1. `ui/src/api/client.ts` — typed async fetch wrapper for all service endpoints.
2. Replace mock imports in `Dashboard.tsx` with live API calls.
3. Replace mock imports in `IncidentDetail.tsx` with live API calls. Wire Approve button to orchestrator.
4. Wire scenario buttons in `FaultInjector.tsx` to fault-injector API.
5. Replace mock imports in `AuditLog.tsx` with live event-store call.
6. Add 10-second polling for Dashboard and KPI panel.
7. Add loading spinner and error banner components.
8. Add incident status badge (detected / diagnosing / remediating / resolved).

When done, start the full stack with `make up` from the repo root, open http://localhost:3000, inject a fault, and verify the incident appears live in the dashboard.
