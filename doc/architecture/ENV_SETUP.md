# Environment Setup Blueprint

## Environments
- `local-dev`: day-to-day development on laptop.
- `local-demo`: stable local profile for customer walkthrough on laptop.

## Core Components
- Local `kind` Kubernetes cluster for sample services.
- Batch worker service with controllable failure hooks.
- Fault injector service or scripts.
- Telemetry stack (metrics, logs, traces, events).
- AI diagnosis service.
- Remediation orchestrator service.
- Web demo console.

## Required Local Dependencies
- Docker Desktop.
- `kind`, `kubectl`, and optional `helm`.
- Local `.env` for service configuration.

## Networking and Access
- Separate namespace per service group.
- Service accounts with least privilege for remediation actions.
- Restricted local ingress to demo UI and APIs.

## Operational Controls
- Feature flags for manual vs auto execution mode.
- Scenario toggles to enable one fault at a time.
- Snapshot/reset script to return to known-good baseline.
- One-command bootstrap script to start full local stack.

## Readiness Checklist
1. All services healthy before demo start.
2. Docker Desktop and `kind` cluster are healthy.
3. Fault injectors tested for each scenario.
4. Approval path verified.
5. KPI capture verified.
6. Rollback/reset command tested.
