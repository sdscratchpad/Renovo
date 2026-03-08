#!/bin/sh
# Full reset: stop all services, wipe the event-store database, then start afresh.
# GITHUB_TOKEN is read from ~/ravi-poc-github-token-github-models,
# falling back to the GITHUB_TOKEN environment variable.
# Usage: ./scripts/restart-services.sh

set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck source=lib.sh
. "$ROOT/scripts/lib.sh"

# Validate token first — fail fast before tearing anything down.
load_github_token || exit 1

stop_all

echo "Wiping event-store database: $DB_PATH"
rm -f "$DB_PATH"
echo "  DB wiped."

start_all
