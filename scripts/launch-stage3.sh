#!/usr/bin/env bash
# launch-stage3.sh
# Opens the full monorepo for Stage 3 — integration and demo hardening.
# Prerequisite: all Stage 1 and Stage 2 services build and run.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo ""
echo "========================================"
echo "  Stage 3 — Integration and Demo Hardening"
echo "========================================"
echo ""
echo "  Prerequisite check:"
echo "  - All 6 services build and start"
echo "  - fault-injector can trigger scenarios"
echo "  - orchestrator can approve and execute"
echo "  - UI shows live data"
echo ""
read -p "  Have all Stage 2 components passed? [y/N] " confirm
if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
  echo "  Aborting. Complete Stage 2 before running Stage 3."
  exit 1
fi
echo ""

echo "Opening full monorepo in VS Code..."
code "$REPO_ROOT"
sleep 2

echo ""
echo "----------------------------------------------"
echo "  ACTION REQUIRED in the VS Code window:"
echo ""
echo "  1. Open Copilot Chat  (⌃⌘I)"
echo "  2. Type:  /agent-integrate"
echo ""
echo "  The agent will:"
echo "  - Wire the root Makefile end-to-end"
echo "  - Run full smoke tests"
echo "  - Tune detection and telemetry"
echo "  - Harden the 10-minute demo path"
echo "----------------------------------------------"
echo ""
