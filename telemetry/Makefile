.PHONY: up down

## Start Prometheus, Grafana, and OTel collector via docker compose
up:
	docker compose up -d
	@echo "✓ telemetry up (Prometheus :9090, Grafana :3001)"

down:
	docker compose down
