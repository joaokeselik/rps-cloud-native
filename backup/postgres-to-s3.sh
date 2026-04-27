#!/bin/sh
set -eu

: "${PGHOST:?PGHOST is required}"
: "${PGDATABASE:?PGDATABASE is required}"
: "${PGUSER:?PGUSER is required}"
: "${PGPASSWORD:?PGPASSWORD is required}"
: "${S3_BUCKET:?S3_BUCKET is required}"

S3_PREFIX="${S3_PREFIX:-postgres-backups}"
BACKUP_NAME="${PGDATABASE}-$(date -u +%Y%m%dT%H%M%SZ).sql.gz"
BACKUP_PATH="/tmp/${BACKUP_NAME}"

echo "Creating PostgreSQL backup ${BACKUP_NAME}"
pg_dump --host "$PGHOST" --username "$PGUSER" --dbname "$PGDATABASE" --format plain | gzip > "$BACKUP_PATH"

echo "Uploading backup to s3://${S3_BUCKET}/${S3_PREFIX}/${BACKUP_NAME}"
aws s3 cp "$BACKUP_PATH" "s3://${S3_BUCKET}/${S3_PREFIX}/${BACKUP_NAME}"

echo "Backup complete"
