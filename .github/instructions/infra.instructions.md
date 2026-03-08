---
applyTo: "infra/**,telemetry/**"
---

# Infra Agent Instructions

## Scope
You own `infra/` and `telemetry/`. Do not modify any service code or UI code.

## What to build
1. `infra/kind-config.yaml` — kind cluster config with 1 control-plane + 2 worker nodes.
2. `infra/k8s/namespaces/` — namespace manifests for `workloads` and `monitoring`.
3. `infra/k8s/sample-app/` — Deployment, Service, HPA manifests for the sample-app.
4. `infra/k8s/batch-worker/` — Job/CronJob manifest for batch-worker.
5. `telemetry/docker-compose.yaml` — Prometheus, Grafana, OTel collector, all pre-configured.
6. `telemetry/prometheus/prometheus.yml` — scrape configs for all services.
7. `telemetry/otel/otel-collector-config.yaml` — receiver/exporter pipeline.
8. `telemetry/grafana/dashboards/` — pre-built dashboard JSON for service health and incidents.

## Key constraints
- Cluster name: `ravi-poc`.
- All telemetry runs via `docker compose`.
- Prometheus scrapes all service `/metrics` endpoints on localhost ports.
- OTel collector listens on `localhost:4317` (gRPC) and `localhost:4318` (HTTP).
- Grafana default credentials: `admin/admin`.

## Done criteria
- `make -C infra up` creates the kind cluster with all namespaces and workload manifests applied.
- `make -C telemetry up` starts Prometheus on :9090, Grafana on :3001, OTel on :4317.
- Prometheus can scrape at least sample-app metrics after `make up` in root.
