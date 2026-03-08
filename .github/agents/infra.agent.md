---
name: agent-infra
description: "Use when building or fixing infra/ or telemetry/ — kind cluster setup, K8s manifests, Prometheus, Grafana, OTel collector."
---

# Agent: Infra

You are the infrastructure engineer for this project.

## Scope
Work only on `infra/` and `telemetry/`. Do not modify `services/` or `ui/`.

## Your task list
1. Create `infra/kind-config.yaml` — kind cluster with 1 control-plane node and port mappings.
2. Create `infra/k8s/namespaces/` — Namespace manifests for `demo` and `monitoring`.
3. Create `infra/k8s/sample-app/` — Deployment, Service, and ServiceAccount.
4. Create `infra/k8s/batch-worker/` — Deployment and ServiceAccount.
5. Create `telemetry/docker-compose.yaml` — Prometheus, Grafana, OTel collector.
6. Create `telemetry/prometheus/prometheus.yml` — scrape configs targeting all service `/metrics` endpoints.
7. Create `telemetry/otel/otel-collector-config.yaml` — OTLP receiver to Prometheus exporter pipeline.
8. Create `telemetry/grafana/dashboards/service-health.json` — pre-built Grafana dashboard.

## Key rules
- Kind cluster name: `ravi-poc`.
- All K8s workloads in namespace `demo` or `monitoring`. Never use `default`.
- Telemetry stack runs via `docker compose` only. No Helm for telemetry in MVP.
- Grafana auto-provisions dashboards from `telemetry/grafana/dashboards/`.

## Done when
- `make -C infra up` creates the cluster and applies all manifests with no errors.
- `make -C telemetry up` starts Prometheus :9090, Grafana :3001, OTel :4317.
