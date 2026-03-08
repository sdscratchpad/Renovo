#!/usr/bin/env bash
# launch-stage2.sh
# Opens 3 VS Code windows for Stage 2 parallel agents.
# Prerequisite: Stage 1 services must be individually functional.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
WORKSPACES="$REPO_ROOT/workspaces"

echo ""
echo "========================================"
echo "  Stage 2 — Parallel Build (3 agents)"
echo "========================================"
echo ""
echo "  Prerequisite check:"
echo "  - sample-app and batch-worker build and start"
echo "  - event-store and diagnosis build and start"
echo "  - UI renders all 4 pages with mock data"
echo ""
read -p "  Have all Stage 1 components passed? [y/N] " confirm
if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
  echo "  Aborting. Complete Stage 1 before running Stage 2."
  exit 1
fi
echo ""

declare -a WINDOWS=(
  "stage2-fault-injector.code-workspace|agent-fault-injector|Window 1 — Fault Injector"
  "stage2-orchestrator.code-workspace|agent-orchestrator|Window 2 — Orchestrator"
  "stage2-ui-wiring.code-workspace|agent-ui-wiring|Window 3 — UI Wiring"
)

for entry in "${WINDOWS[@]}"; do
  workspace="${entry%%|*}"
  rest="${entry#*|}"
  prompt="${rest%%|*}"
  label="${rest#*|}"

  echo "Opening: $label"
  code --new-window "$WORKSPACES/$workspace"
  sleep 1
done

echo ""
echo "----------------------------------------------"
echo "  3 windows are now open."
echo "  In each window, open Copilot Chat and type:"
echo ""
echo "  Window 1 (fault-injector):  /agent-fault-injector"
echo "  Window 2 (orchestrator):    /agent-orchestrator"
echo "  Window 3 (ui-wiring):       /agent-ui-wiring"
echo "----------------------------------------------"
echo ""
