CLUSTER_NAME=ravi-poc
KUBECONFIG_PATH=$(HOME)/.kube/config

.PHONY: up down reset kind-load

## Create local kind cluster and apply base manifests
up:
	@if kind get clusters 2>/dev/null | grep -q $(CLUSTER_NAME); then \
		echo "kind cluster $(CLUSTER_NAME) already running"; \
	else \
		kind create cluster --name $(CLUSTER_NAME) --config kind-config.yaml; \
	fi
	kubectl apply -f k8s/namespaces/
	kubectl apply -f k8s/sample-app/
	kubectl apply -f k8s/batch-worker/
	@echo "✓ infra up"

## Delete the kind cluster
down:
	kind delete cluster --name $(CLUSTER_NAME) || true
	@echo "✓ infra down"
## Build workload images and load them into the kind cluster (run from repo root)
kind-load:
	@echo "Building and loading workload images into kind cluster $(CLUSTER_NAME)..."
	docker build -f ../services/sample-app/Dockerfile  -t sample-app:latest  ..
	docker build -f ../services/batch-worker/Dockerfile -t batch-worker:latest ..
	kind load docker-image sample-app:latest  --name $(CLUSTER_NAME)
	kind load docker-image batch-worker:latest --name $(CLUSTER_NAME)
	kubectl rollout restart deployment/sample-app -n workloads
	kubectl rollout restart cronjob/batch-worker  -n workloads 2>/dev/null || true
	@echo "✓ images loaded and workloads restarted"

## Reset cluster to clean baseline
reset: down up
	@echo "✓ infra reset"
