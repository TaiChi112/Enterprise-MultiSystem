#!/usr/bin/env bash
set -euo pipefail

# One-command KPI seed refresh without dropping docker volumes.
# Prerequisite: docker compose services are up and postgres is running.

cd "$(dirname "$0")/.."

echo "[1/3] Ensuring postgres container is up..."
docker compose up -d postgres >/dev/null

echo "[2/3] Applying KPI DDL (idempotent)..."
docker compose exec -T postgres psql -U postgres -d pos_wms -f /docker-entrypoint-initdb.d/02-kpi-ddl.sql >/dev/null

echo "[3/3] Refreshing KPI seed data..."
docker compose exec -T postgres psql -U postgres -d pos_wms -f /docker-entrypoint-initdb.d/03-kpi-seed.sql >/dev/null

echo "KPI seed refresh completed successfully."
