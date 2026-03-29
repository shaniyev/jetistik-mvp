#!/usr/bin/env bash
set -euo pipefail

# Restore a production database dump and media files into the local dev
# environment (docker compose services: db, minio).
#
# Prerequisites:
#   - docker compose services running (make up)
#   - data/prod-dump.sql exists (run make sync-prod first)
#   - mc (MinIO Client) installed: brew install minio/stable/mc

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
DATA_DIR="$ROOT_DIR/data"
DUMP_FILE="$DATA_DIR/prod-dump.sql"
MEDIA_DIR="$DATA_DIR/media"

# Local DB connection (matches docker-compose.yml)
DB_CONTAINER="$(docker compose -f "$ROOT_DIR/docker-compose.yml" ps -q db 2>/dev/null || true)"
MINIO_ALIAS="jetistik-local"
MINIO_BUCKET="media"
MINIO_ENDPOINT="http://localhost:9000"
MINIO_USER="minioadmin"
MINIO_PASS="minioadmin"

# --- Preflight checks --------------------------------------------------------
if [ ! -f "$DUMP_FILE" ]; then
  echo "ERROR: $DUMP_FILE not found. Run 'make sync-prod' first."
  exit 1
fi

if [ -z "$DB_CONTAINER" ]; then
  echo "ERROR: db container is not running. Run 'make up' first."
  exit 1
fi

echo "==> Local data dir: $DATA_DIR"
echo ""

# --- Restore database --------------------------------------------------------
echo "==> Dropping and recreating database..."
docker compose -f "$ROOT_DIR/docker-compose.yml" exec -T db \
  psql -U jetistik -d postgres -c "DROP DATABASE IF EXISTS jetistik;"
docker compose -f "$ROOT_DIR/docker-compose.yml" exec -T db \
  psql -U jetistik -d postgres -c "CREATE DATABASE jetistik OWNER jetistik;"

echo "==> Restoring database from dump..."
docker compose -f "$ROOT_DIR/docker-compose.yml" exec -T db \
  psql -U jetistik -d jetistik < "$DUMP_FILE"

echo "    Database restored."
echo ""

# --- Restore media files to MinIO --------------------------------------------
if [ -d "$MEDIA_DIR" ] && [ "$(ls -A "$MEDIA_DIR" 2>/dev/null)" ]; then
  if command -v mc &>/dev/null; then
    echo "==> Configuring MinIO client..."
    mc alias set "$MINIO_ALIAS" "$MINIO_ENDPOINT" "$MINIO_USER" "$MINIO_PASS" --api S3v4 >/dev/null

    echo "==> Creating bucket '$MINIO_BUCKET' (if not exists)..."
    mc mb --ignore-existing "$MINIO_ALIAS/$MINIO_BUCKET"

    echo "==> Uploading media files to MinIO..."
    mc mirror --overwrite "$MEDIA_DIR/" "$MINIO_ALIAS/$MINIO_BUCKET/"

    FILE_COUNT=$(find "$MEDIA_DIR" -type f | wc -l | tr -d ' ')
    echo "    Uploaded $FILE_COUNT files to MinIO bucket '$MINIO_BUCKET'."
  else
    echo "WARNING: 'mc' (MinIO Client) not found."
    echo "  Install it:  brew install minio/stable/mc"
    echo "  Media files are available at: $MEDIA_DIR"
    echo "  You can upload them manually or mount the directory."
  fi
else
  echo "==> No media files to restore (directory empty or missing)."
fi

echo ""
echo "==> Restore complete."
