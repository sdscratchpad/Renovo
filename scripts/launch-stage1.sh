#!/usr/bin/env bash
# launch-stage1.sh
# Opens 4 VS Code windows for Stage 1 parallel agents.
# Each window is scoped to its domain via a multi-root workspace file.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
WORKSPACES="$REPO_ROOT/workspaces"

echo ""
echo "========================================"
echo "  Stage 1 — Parallel Build (4 agents)"
echo "========================================"
echo ""

declare -a WINDOWS=(
  "stage1-infra.code-workspace|agent-infra|Window 1 — Infra (infra/ + telemetry/)"
  "stage1-workloads.code-workspace|agent-workloads|Window 2 — Workloads (sample-app + batch-worker)"
  "stage1-brain.code-workspace|agent-brain|Window 3 — Brain (diagnosis + event-store)"
  "stage1-ui.code-workspace|agent-ui|Window 4 — UI (ui/)"
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
echo "  4 windows are now open."
echo "  In each window, open Copilot Chat and type:"
echo ""
echo "  Window 1 (infra):      /agent-infra"
echo "  Window 2 (workloads):  /agent-workloads"
echo "  Window 3 (brain):      /agent-brain"
echo "  Window 4 (ui):         /agent-ui"
echo ""
echo "  All 4 agents work concurrently with no overlap."
echo "  Ensure GITHUB_TOKEN is set in your shell before"
echo "  starting the brain agent (diagnosis needs it)."
echo "----------------------------------------------"
echo ""
