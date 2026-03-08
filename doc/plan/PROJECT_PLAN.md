# Project Plan (MVP)

## Goal
Deliver a customer-ready demo of an AI Infrastructure Resilience Copilot that can detect, diagnose, and remediate selected cloud incidents in a controlled environment.
The full MVP must run from a laptop with local infrastructure and product services.

## Scope Lock (MVP)
- Include only 3 incident scenarios for v1.
- Include Kubernetes + one batch path.
- Use human approval by default for remediation.
- Run infrastructure on local `kind` cluster for both development and demo.
- Run product services locally (containerized) without cloud dependency.
- Exclude multi-cloud and full autonomous production operations.

## Timeline (4-6 Weeks)

### Week 1: Foundation
- Provision local demo environment (`kind` cluster, sample app, batch worker).
- Set up telemetry pipeline (metrics, logs, traces).
- Implement baseline dashboard for health and incidents.
- Add one-command local bootstrap and reset scripts.

### Week 2: Fault Lab + Detection
- Implement fault injectors for 3 MVP scenarios.
- Build anomaly detection rules and event pipeline.
- Add incident timeline in demo UI.

### Week 3: AI Diagnosis
- Implement RCA summarization service with confidence scoring.
- Map evidence sources (metrics/log signatures/events) to each scenario.
- Show RCA output in UI.

### Week 4: Remediation Orchestration
- Implement runbook execution engine.
- Add policy checks and blast-radius controls.
- Integrate manual approval workflow in UI.

### Week 5: Validation + Demo Hardening
- Run full demo dry-runs and tune thresholds.
- Add KPI tracking (MTTD, MTTR, auto-resolve rate).
- Finalize demo script and fallback paths.

### Week 6 (Optional Buffer)
- Add low-risk auto mode for one scenario.
- Improve UX polish and prepare customer Q&A appendix.

## Team and Ownership
- Platform Engineer: environment, observability, fault tooling.
- Backend Engineer: diagnosis API, runbook orchestrator.
- Frontend Engineer: demo UI, timeline, KPI views.
- Tech Lead: scope control, acceptance sign-off, demo rehearsal.

## Exit Criteria
- All 3 scenarios can be demonstrated end-to-end.
- Manual remediation path works reliably.
- KPI deltas are captured and shown in UI.
- Demo completes in 10 minutes with repeatable outcome.
- System can be started from a laptop and reset to baseline in predictable time.
