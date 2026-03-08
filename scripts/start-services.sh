#!/bin/sh
# Start all backend services (non-UI) in the background.
# GITHUB_TOKEN is read from ~/ravi-poc-github-token-github-models,
# falling back to the GITHUB_TOKEN environment variable.
# Usage: ./scripts/start-services.sh

set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck source=lib.sh
. "$ROOT/scripts/lib.sh"

load_github_token || exit 1

stop_all
start_all
