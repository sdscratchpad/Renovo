#!/bin/sh
# Show running status and log file info for all backend services.
# Usage: ./scripts/services-status.sh

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck source=lib.sh
. "$ROOT/scripts/lib.sh"

# check_service <name> <port>
check_service() {
  name="$1"
  port="$2"
  log="/tmp/${name}.log"

  pid="$(pgrep -f "bin/${name}" 2>/dev/null | head -1)"
  if [ -n "$pid" ]; then
    status="RUNNING (PID $pid)"
  else
    status="STOPPED"
  fi

  if [ -f "$log" ]; then
    size="$(wc -c < "$log" | tr -d ' ') bytes"
  else
    size="not found"
  fi

  printf "  %-18s :%-6s %s\n" "$name" "$port" "$status"
  printf "    log: %s (%s)\n" "$log" "$size"
}

# check_ui — special variant for the React UI (process: react-scripts start).
check_ui() {
  port=3000
  log="/tmp/ui.log"

  pid="$(pgrep -f "react-scripts start" 2>/dev/null | head -1)"
  if [ -n "$pid" ]; then
    status="RUNNING (PID $pid)"
  else
    status="STOPPED"
  fi

  if [ -f "$log" ]; then
    size="$(wc -c < "$log" | tr -d ' ') bytes"
  else
    size="not found"
  fi

  printf "  %-18s :%-6s %s\n" "ui" "$port" "$status"
  printf "    log: %s (%s)\n" "$log" "$size"
}

echo "Service status — $(date)"
echo "-----------------------------------------------------------"
check_service "event-store"    8085
check_service "diagnosis"      8083
check_service "orchestrator"   8084
check_service "fault-injector" 8082
check_ui

echo ""
echo "Token file:"
if [ -f "$TOKEN_FILE" ]; then
  echo "  PRESENT  $TOKEN_FILE"
else
  echo "  MISSING  $TOKEN_FILE"
fi

echo ""
echo "Event-store DB:"
if [ -f "$DB_PATH" ]; then
  size="$(wc -c < "$DB_PATH" | tr -d ' ') bytes"
  echo "  PRESENT  $DB_PATH ($size)"
else
  echo "  ABSENT   $DB_PATH"
fi
