# Technical Stack (MVP Recommendation)

## Runtime Model
- Primary mode: local laptop development and demo.
- Orchestration: `kind` Kubernetes cluster.
- Container runtime: Docker Desktop.
- Compute: Kubernetes services + containerized batch worker.

## Local Tooling
- Cluster management: `kind` + `kubectl`.
- Packaging/deploy: `helm` or raw manifests for MVP.
- Dev automation: Makefile or task runner for bootstrap/reset.

## Telemetry and Observability
- Metrics: Prometheus.
- Logs/Traces: OpenTelemetry collector pipeline.
- Visualization: Grafana + custom demo UI panels.

## AI and Decision Layer
- LLM service for RCA summarization and action explanation.
- Rule/runbook engine for deterministic and safe execution.
- Confidence scoring + risk labels for each recommendation.

## Remediation Layer
- Runbook executor (API-driven) for restart, rollback, scale, and retry actions.
- Policy engine for allow-list and blast-radius limits.

## Demo UI
- Web app with incident timeline, RCA panel, remediation actions, and KPI dashboard.

## Data and Audit
- Incident/event store for timeline history.
- Action audit log for approvals and execution outcomes.

## Portability Note
- Cloud deployment is a later phase.
- MVP proof is local-first to keep demos repeatable and independent of cloud setup.
