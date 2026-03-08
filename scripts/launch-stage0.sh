#!/usr/bin/env bash
# launch-stage0.sh
# Opens a single VS Code window with the full monorepo for Stage 0 scaffold validation.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo ""
echo "========================================"
echo "  Stage 0 — Foundation and Contracts"
echo "========================================"
echo ""
echo "Opening full monorepo in VS Code..."
code "$REPO_ROOT"

sleep 2

echo ""
echo "----------------------------------------------"
echo "  ACTION REQUIRED in the VS Code window:"
echo ""
echo "  1. Open Copilot Chat  (⌃⌘I or View → Copilot Chat)"
echo "  2. Type:  /agent-setup"
echo "  3. This will validate the scaffold and"
echo "     report any gaps before Stage 1 begins."
echo "----------------------------------------------"
echo ""
