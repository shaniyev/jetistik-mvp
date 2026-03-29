#!/usr/bin/env bash
set -euo pipefail

# Download production database dump and media files from the Jetistik v1 server.
# Usage: ./scripts/sync-prod-data.sh [SSH_USER@]HOST
#
# Defaults come from the Ansible inventory (almalinux@213.155.23.19).
# The production app lives at /opt/jetistik and runs via docker compose.

PROD_HOST="${1:-almalinux@213.155.23.19}"
APP_DIR="/opt/jetistik"
DATA_DIR="$(cd "$(dirname "$0")/.." && pwd)/data"

mkdir -p "$DATA_DIR/media"

echo "==> Target server: $PROD_HOST"
echo "==> Local data dir: $DATA_DIR"
echo ""

# --- Database dump -----------------------------------------------------------
echo "==> Dumping production database..."
ssh "$PROD_HOST" \
  "cd $APP_DIR && docker compose exec -T db pg_dump -U jetistik --clean --if-exists jetistik" \
  > "$DATA_DIR/prod-dump.sql"

DUMP_SIZE=$(wc -c < "$DATA_DIR/prod-dump.sql" | tr -d ' ')
echo "    Saved $DATA_DIR/prod-dump.sql (${DUMP_SIZE} bytes)"

# --- Media files -------------------------------------------------------------
echo "==> Packaging media files on server..."
ssh "$PROD_HOST" \
  "cd $APP_DIR && docker compose exec -T web tar czf - -C /app/media ." \
  > "$DATA_DIR/media.tar.gz"

echo "==> Extracting media files locally..."
tar xzf "$DATA_DIR/media.tar.gz" -C "$DATA_DIR/media"
rm -f "$DATA_DIR/media.tar.gz"

MEDIA_COUNT=$(find "$DATA_DIR/media" -type f | wc -l | tr -d ' ')
echo "    Extracted $MEDIA_COUNT files into $DATA_DIR/media/"

echo ""
echo "==> Sync complete. Run 'make restore-data' to load into your local environment."
