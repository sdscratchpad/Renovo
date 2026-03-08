# Demo Script: AI Infrastructure Resilience Copilot

## Objective
Show that AI can detect, explain, and safely remediate cloud infrastructure incidents with measurable impact.

## Local Demo Assumption
- Entire demo runs on a laptop.
- Infrastructure runs on local `kind` Kubernetes.
- Product services run locally in containers.

## Runtime Modes
- Manual mode: operator approves remediation.
- Auto mode: only low-risk, policy-approved actions execute automatically.

## 10-Minute Demo Sequence
1. Baseline health
- Open Web Demo Console.
- Confirm `kind` cluster and all local services are healthy.
- Show all services healthy and KPI baseline.

2. Inject fault
- Trigger one scenario (for example bad deployment config).
- Show service health degradation.

3. AI diagnosis
- Show anomaly detection event.
- Show AI RCA summary with evidence and confidence.

4. Remediation proposal
- Show recommended action and risk level.
- In manual mode, approve action.

5. Recovery
- Execute remediation.
- Show service recovery and KPI delta (MTTD, MTTR).

6. Close
- Show audit trail of what happened.
- Reinforce guardrails and business value.

## Pre-Demo Checklist (Local)
1. Docker Desktop is running.
2. `kind` cluster is created and reachable.
3. Product services are up and connected to telemetry.
4. Fault injector is enabled for selected scenario.
5. Reset command is tested before customer session.

## Suggested Fault Order
1. Bad rollout in Kubernetes.
2. Resource saturation event.
3. Failed batch retry scenario.

## Key Messages to Say During Demo
- The AI is constrained by policies and runbooks.
- Human approval remains in control by default.
- The outcome is faster recovery and lower downtime cost.

## Fallback Procedures

### If a service health check fails before demo
```bash
export GITHUB_TOKEN=$(cat ~/ravi-poc-github-token-github-models)
make reset   # fast reset: wipes DB, restores K8s, restarts services
```

### If fault injection returns an error
Check the kind cluster is reachable:
```bash
kubectl get nodes
# If cluster is down:
make cluster-reset   # ~5 min, recreates the cluster
```

### If AI diagnosis times out or returns empty RCA
The diagnosis service requires GITHUB_TOKEN. Verify it is set:
```bash
curl http://localhost:8083/health
# If 404/connection refused, restart:
pkill -f bin/diagnosis
export GITHUB_TOKEN=$(cat ~/ravi-poc-github-token-github-models)
cd services/diagnosis && make up
```

### If KPI panel shows 0 values
KPI data is written after diagnosis + remediation complete. Trigger the full flow:
1. Inject a fault: `curl -X POST http://localhost:8082/inject/bad-rollout`
2. Run diagnosis: open the UI incident detail and wait for RCA
3. Approve the remediation: click Approve in the UI
4. Refresh — MTTD and MTTR should populate within 5 seconds.

### If orchestrator returns "namespace not in allow-list"
The binary may be stale. Rebuild:
```bash
cd services/orchestrator && make build && make down && make up
```

### Manual scenario restore (if UI approval is not available)
```bash
# Undo all injected faults at once:
curl -X POST http://localhost:8082/restore

# Or per scenario:
curl -X POST http://localhost:8082/restore/bad-rollout
curl -X POST http://localhost:8082/restore/resource-saturation
curl -X POST http://localhost:8082/restore/batch-timeout
```

### Full environment restart order
```bash
export GITHUB_TOKEN=$(cat ~/ravi-poc-github-token-github-models)
make reset
# Then open http://localhost:3000
```
