# Demo Acceptance Criteria

## Mandatory (Must Pass)
1. Entire stack runs on a laptop with local `kind` Kubernetes.
2. Scenario can be injected on demand.
3. Incident is detected and shown in UI within target detection window.
4. RCA is generated with evidence and confidence.
5. Recommended remediation is shown with risk level.
6. Manual approval triggers remediation successfully.
7. Recovery is verified and health returns to normal state.
8. KPI delta is displayed for MTTD and MTTR.
9. Audit trail records incident, recommendation, approval, and action result.
10. Local environment reset returns to known-good state before the next run.

## Reliability Targets (MVP)
- End-to-end demo success rate: >= 90% in internal dry runs.
- Detection latency target: <= 60 seconds for selected scenarios.
- Recovery completion target: <= 5 minutes for selected scenarios.
- Full local stack startup target: <= 15 minutes on agreed laptop profile.

## Non-Goals for MVP
- No broad autonomous remediation across all workloads.
- No multi-region or multi-cloud orchestration.
- No production rollout commitments from demo environment.
