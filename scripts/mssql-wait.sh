#!/usr/bin/env bash
set -euo pipefail

CONTAINER_NAME="${CONTAINER_NAME:-sqlserver}"
MSSQL_SA_PASSWORD="${MSSQL_SA_PASSWORD:-Passw0rd123}"
TIMEOUT_SECONDS="${TIMEOUT_SECONDS:-120}"
SLEEP_SECONDS="${SLEEP_SECONDS:-2}"
SQLCMD="/opt/mssql-tools18/bin/sqlcmd"

elapsed=0

echo "Waiting for SQL Server in container: $CONTAINER_NAME"

while true; do
  if docker exec "$CONTAINER_NAME" "$SQLCMD" -S localhost -U sa -P "$MSSQL_SA_PASSWORD" -C -Q "SELECT 1" -b >/dev/null 2>&1; then
    echo "SQL Server is ready."
    exit 0
  fi

  if (( elapsed >= TIMEOUT_SECONDS )); then
    echo "Timed out waiting for SQL Server readiness after ${TIMEOUT_SECONDS}s." >&2
    exit 1
  fi

  sleep "$SLEEP_SECONDS"
  elapsed=$((elapsed + SLEEP_SECONDS))
done
