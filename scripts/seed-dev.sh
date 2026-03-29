#!/usr/bin/env bash
set -euo pipefail

# Pull production data and load it into the local dev environment in one step.
# Usage: ./scripts/seed-dev.sh [SSH_USER@HOST]

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "========================================"
echo "  Jetistik: Seed local dev from prod"
echo "========================================"
echo ""

if [ -n "${1:-}" ]; then
  "$SCRIPT_DIR/sync-prod-data.sh" "$1"
else
  "$SCRIPT_DIR/sync-prod-data.sh"
fi
echo ""
"$SCRIPT_DIR/restore-prod-data.sh"

echo ""
echo "========================================"
echo "  Done. Local environment is seeded."
echo "========================================"
