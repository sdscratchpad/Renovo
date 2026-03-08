# MVP Backlog

## Must-Have
- Build laptop-first sandbox with one local `kind` Kubernetes cluster and one batch worker.
- Add one-command local bootstrap script for product + infrastructure.
- Add one-command local reset script to restore baseline between demo runs.
- Implement 3 fault scenarios:
  - Bad deployment rollout.
  - Resource saturation.
  - Batch dependency timeout.
- Ingest metrics, logs, traces, and events into unified timeline.
- Implement anomaly detection and incident trigger service.
- Implement RCA service with evidence mapping and confidence score.
- Implement runbook orchestrator with manual approval gate.
- Implement policy guardrails and blast-radius checks.
- Build demo UI: health map, incident timeline, RCA panel, action panel, KPI panel.
- Capture and display MTTD and MTTR before/after values.
- Rehearse 10-minute demo path and fallback path.
- Validate demo on target laptop hardware profile.

## Nice-to-Have
- Auto-remediation mode for one low-risk scenario.
- What-if simulation mode for remediation action.
- AI-generated post-incident report export.
- Role-based view for operator vs executive audience.
