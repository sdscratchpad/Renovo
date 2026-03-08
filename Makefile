.PHONY: help infra-up up down reset cluster-reset demo build kind-load test

## Show this help message
help:
	@echo "Usage: make <target>"
	@echo ""
	@awk '/^## /{if(desc=="")desc=substr($$0,4); next} /^[a-zA-Z_-]+:/{printf "  %-16s %s\n", substr($$1,1,length($$1)-1), desc; desc=""}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

## Start kind cluster and telemetry stack only
infra-up:
	$(MAKE) -C infra up
	$(MAKE) -C telemetry up
	@echo "✓ Infra started. Prometheus: http://localhost:9090  Grafana: http://localhost:3001"

## Start all components (infra first, then services, then UI)
up:
	$(MAKE) -C infra up
	$(MAKE) -C telemetry up
	$(MAKE) -C services/event-store up
	$(MAKE) -C services/diagnosis up
	$(MAKE) -C services/orchestrator up
	$(MAKE) -C services/sample-app up
	$(MAKE) -C services/batch-worker up
	$(MAKE) -C services/fault-injector up
	$(MAKE) -C ui up
	@echo "✓ All components started. Open http://localhost:3000 for the demo console."

## Stop all components
down:
	$(MAKE) -C ui down
	$(MAKE) -C services/fault-injector down
	$(MAKE) -C services/batch-worker down
	$(MAKE) -C services/sample-app down
	$(MAKE) -C services/orchestrator down
	$(MAKE) -C services/diagnosis down
	$(MAKE) -C services/event-store down
	$(MAKE) -C telemetry down
	$(MAKE) -C infra down

## Fast demo reset: stop services, wipe event data, restore K8s baseline, restart.
## Does NOT recreate the kind cluster. Export GITHUB_TOKEN before running.
reset:
	@echo "--- Stopping product services ---"
	-$(MAKE) -C ui down
	-$(MAKE) -C services/fault-injector down
	-$(MAKE) -C services/batch-worker down
	-$(MAKE) -C services/sample-app down
	-$(MAKE) -C services/orchestrator down
	-$(MAKE) -C services/diagnosis down
	-$(MAKE) -C services/event-store down
	@echo "--- Wiping event store database ---"
	rm -f services/event-store/data/events.db
	@echo "--- Restoring Kubernetes workloads to baseline ---"
	-kubectl rollout undo deployment/sample-app -n workloads
	-kubectl patch deployment/sample-app -n workloads --type=strategic -p '{"spec":{"template":{"spec":{"containers":[{"name":"sample-app","resources":{"limits":{"cpu":"500m","memory":"128Mi"},"requests":{"cpu":"50m","memory":"64Mi"}}}]}}}}'
	-kubectl set env cronjob/batch-worker -n workloads FAIL_MODE-
	@echo "--- Restarting product services ---"
	$(MAKE) -C services/event-store up
	$(MAKE) -C services/diagnosis up
	$(MAKE) -C services/orchestrator up
	$(MAKE) -C services/fault-injector up
	$(MAKE) -C services/sample-app up
	$(MAKE) -C services/batch-worker up
	@printf "\u2713 Reset complete. Open http://localhost:3000 for a fresh demo.\n"

## Full cluster reset: destroys and recreates the kind cluster (~5 min).
cluster-reset: down
	$(MAKE) -C infra reset
	$(MAKE) up

## Run full demo dry-run: up + inject first scenario + print status
demo: up
	@echo "Triggering scenario: bad-rollout"
	curl -sf -X POST http://localhost:8082/inject/bad-rollout || echo "fault-injector not ready yet"

## Build all Go services
build:
	$(MAKE) -C services/sample-app build
	$(MAKE) -C services/batch-worker build
	$(MAKE) -C services/fault-injector build
	$(MAKE) -C services/diagnosis build
	$(MAKE) -C services/orchestrator build
	$(MAKE) -C services/event-store build

## Build workload Docker images and load into the kind cluster
kind-load:
	$(MAKE) -C infra kind-load

## Run all tests
test:
	$(MAKE) -C services/sample-app test
	$(MAKE) -C services/batch-worker test
	$(MAKE) -C services/fault-injector test
	$(MAKE) -C services/diagnosis test
	$(MAKE) -C services/orchestrator test
	$(MAKE) -C services/event-store test
