# AI Infrastructure Resilience Copilot — Workspace Instructions

## Project purpose
This monorepo contains a demo product that detects, diagnoses, and remediates cloud infrastructure
incidents using AI. It is built to run entirely on a laptop using a local `kind` Kubernetes cluster.

## Monorepo layout
```
ravi-poc/
├── contracts/            shared Go types (IncidentEvent, RCAPayload, RemediationAction, ...)
├── infra/                kind cluster config and Kubernetes manifests
├── telemetry/            Prometheus, Grafana, OTel collector (docker compose)
├── services/
│   ├── sample-app/       Go HTTP service — monitored workload
│   ├── batch-worker/     Go recurring batch job — controllable failure hooks
│   ├── fault-injector/   Go HTTP API — triggers MVP fault scenarios on demand
│   ├── diagnosis/        Go service — AI RCA engine using GitHub Models GPT-4o
│   ├── orchestrator/     Go service — runbook executor, Kubernetes remediation actions
│   └── event-store/      Go service — SQLite persistence for all incident/audit data
├── ui/                   React + TypeScript demo console
├── workspaces/           VS Code multi-root workspace files (one per parallel agent)
└── scripts/              Launch scripts to open all agent windows
```

## Stack conventions
- **Backend**: Go 1.23. Standard library HTTP where possible. No frameworks unless necessary.
- **Frontend**: React 18 + TypeScript. No CSS-in-JS; use plain CSS modules.
- **Local infra**: `kind` cluster (`ravi-poc`). Manifests in `infra/k8s/`.
- **Telemetry**: Prometheus metrics, OpenTelemetry traces, structured JSON logs.
- **AI**: GPT-4o via GitHub Models. Endpoint: `https://models.inference.ai.azure.com`. Auth: `GITHUB_TOKEN` env var (PAT with `models:read` scope — no Azure account needed).
- **Persistence**: SQLite via `go-sqlite3` in event-store only. No external databases.

## GitHub Models auth pattern
```go
// Use this pattern in the diagnosis service for all LLM calls.
import "net/http"

const githubModelsEndpoint = "https://models.inference.ai.azure.com/chat/completions"
const model = "gpt-4o"

func newLLMRequest(token, prompt string) *http.Request {
    // POST to githubModelsEndpoint with Authorization: Bearer <token>
    // Body: {"model": "gpt-4o", "messages": [...], "temperature": 0.2}
}
```

## Service ports (local)
| Service        | Port |
|----------------|------|
| sample-app     | 8080 |
| batch-worker   | —    |
| fault-injector | 8082 |
| diagnosis      | 8083 |
| orchestrator   | 8084 |
| event-store    | 8085 |
| ui             | 3000 |
| Prometheus     | 9090 |
| Grafana        | 3001 |

## Inter-service communication (local)
All services communicate over localhost. Use `http://localhost:<port>` in dev.

## Shared types
All event/payload types live in `contracts/types.go`. Import as:
```go
import "github.com/ravi-poc/contracts"
```
Each service's `go.mod` has a `replace` directive pointing to `../../contracts`.

## MVP fault scenarios
1. `bad-rollout` — Deploy a misconfigured version of sample-app; AI detects error rate spike; orchestrator rolls back.
2. `resource-saturation` — Inject CPU stress; AI detects saturation; orchestrator scales up replicas.
3. `batch-timeout` — Block the batch-worker dependency; AI detects job stall; orchestrator retries.

## Coding conventions
- Every HTTP handler returns JSON. Content-Type: application/json.
- All errors are logged with `log.Printf("component: %v", err)`.
- TODO comments use the format `// TODO(agent-<name>): description` to signal the owning agent.
- No global state. Pass dependencies via struct fields or function parameters.
- Unit tests go in `_test.go` files alongside the code they test.
