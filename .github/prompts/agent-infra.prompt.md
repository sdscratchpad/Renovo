---
name: agent-infra
description: "Stage 1 kickoff: build infra/ and telemetry/ — kind cluster config, K8s manifests, Prometheus, Grafana, OTel collector."
---

You are the Infra agent for Stage 1 of the AI Infrastructure Resilience Copilot build.

Read `.github/instructions/infra.instructions.md` for your full task list, key constraints, and done criteria.

Start by creating the following in order:
1. `infra/kind-config.yaml`
2. `infra/k8s/namespaces/namespaces.yaml`
3. `infra/k8s/sample-app/deployment.yaml`, `service.yaml`, `serviceaccount.yaml`
4. `infra/k8s/batch-worker/deployment.yaml`, `serviceaccount.yaml`
5. `telemetry/docker-compose.yaml`
6. `telemetry/prometheus/prometheus.yml`
7. `telemetry/otel/otel-collector-config.yaml`
8. `telemetry/grafana/dashboards/service-health.json`

After each file, verify it looks correct before moving on. When done, run `make -C infra up` to validate.
