#!/bin/sh
# lib.sh — shared library sourced by start-services.sh, restart-services.sh,
# and services-status.sh.
#
# Requires: ROOT must be set by the calling script before sourcing this file.

TOKEN_FILE="$HOME/ravi-poc-github-token-github-models"
DB_PATH="$ROOT/services/event-store/data/events.db"

# load_github_token — reads the token from TOKEN_FILE (trimming whitespace),
# or falls back to the GITHUB_TOKEN env var.  Prints a one-line OK/ERROR
# message and returns 1 on failure.
load_github_token() {
  if [ -f "$TOKEN_FILE" ]; then
    token="$(tr -d '[:space:]' < "$TOKEN_FILE")"
    if [ -z "$token" ]; then
      echo "ERROR: token file exists but is empty: $TOKEN_FILE" >&2
      return 1
    fi
    export GITHUB_TOKEN="$token"
    echo "OK: GITHUB_TOKEN loaded from $TOKEN_FILE"
    return 0
  fi

  if [ -n "$GITHUB_TOKEN" ]; then
    echo "OK: GITHUB_TOKEN loaded from environment"
    return 0
  fi

  echo "ERROR: GITHUB_TOKEN not found." >&2
  echo "  Create $TOKEN_FILE with your GitHub PAT (models:read scope)," >&2
  echo "  or export GITHUB_TOKEN=<token> before running this script." >&2
  return 1
}

# stop_all — sends SIGTERM to all backend service and UI processes.
stop_all() {
  echo "Stopping services..."
  pkill -f "bin/event-store"    2>/dev/null || true
  pkill -f "bin/diagnosis"      2>/dev/null || true
  pkill -f "bin/orchestrator"   2>/dev/null || true
  pkill -f "bin/fault-injector" 2>/dev/null || true
  pkill -f "bin/batch-worker"   2>/dev/null || true
  pkill -f "bin/sample-app"     2>/dev/null || true
  pkill -f "react-scripts start" 2>/dev/null || true
  sleep 1
}

# start_all — launches all four backend services in the background using nohup.
# Assumes GITHUB_TOKEN is already exported.
start_all() {
  echo "Starting event-store on :8085..."
  nohup env DB_PATH="$DB_PATH" \
    "$ROOT/services/event-store/bin/event-store" \
    > /tmp/event-store.log 2>&1 &
  echo "  PID $!"

  sleep 1

  echo "Starting diagnosis on :8083..."
  nohup env GITHUB_TOKEN="$GITHUB_TOKEN" \
    "$ROOT/services/diagnosis/bin/diagnosis" \
    > /tmp/diagnosis.log 2>&1 &
  echo "  PID $!"

  echo "Starting orchestrator on :8084..."
  nohup "$ROOT/services/orchestrator/bin/orchestrator" \
    > /tmp/orchestrator.log 2>&1 &
  echo "  PID $!"

  echo "Starting fault-injector on :8082..."
  nohup "$ROOT/services/fault-injector/bin/fault-injector" \
    > /tmp/fault-injector.log 2>&1 &
  echo "  PID $!"

  echo "Starting batch-worker (metrics on :9091)..."
  nohup "$ROOT/services/batch-worker/bin/batch-worker" \
    > /tmp/batch-worker.log 2>&1 &
  echo "  PID $!"

  echo "Starting sample-app on :8080..."
  nohup "$ROOT/services/sample-app/bin/sample-app" \
    > /tmp/sample-app.log 2>&1 &
  echo "  PID $!"

  echo "Starting ui on :3000..."
  nohup sh -c "cd '$ROOT/ui' && PORT=3000 npm start" \
    > /tmp/ui.log 2>&1 &
  echo "  PID $!"

  health_check_all
}

# _wait_for_health <port> — retries /health up to 8 times (1 s apart).
_wait_for_health() {
  port="$1"
  i=0
  while [ "$i" -lt 8 ]; do
    result="$(curl -sf --max-time 2 "http://localhost:$port/health" 2>/dev/null)"
    if [ -n "$result" ]; then
      echo "$result"
      return 0
    fi
    i=$((i + 1))
    sleep 1
  done
  echo "UNREACHABLE"
}

# health_check_all — waits for every service to respond on /health (UI on /).
health_check_all() {
  echo ""
  echo "Health checks:"
  for port in 8085 8083 8084 8082; do
    printf "  :%s -> " "$port"
    _wait_for_health "$port"
  done
  printf "  :3000 -> "
  _wait_for_health_url "http://localhost:3000"
}

# _wait_for_health_url <url> — retries a plain URL up to 8 times (1 s apart).
_wait_for_health_url() {
  url="$1"
  i=0
  while [ "$i" -lt 8 ]; do
    if curl -sf --max-time 2 "$url" > /dev/null 2>&1; then
      echo "OK"
      return 0
    fi
    i=$((i + 1))
    sleep 1
  done
  echo "UNREACHABLE"
}
