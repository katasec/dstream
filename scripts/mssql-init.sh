#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONTAINER_NAME="${CONTAINER_NAME:-sqlserver}"
MSSQL_SA_PASSWORD="${MSSQL_SA_PASSWORD:-Passw0rd123}"
SQLCMD="/opt/mssql-tools18/bin/sqlcmd"

"$SCRIPT_DIR/mssql-start.sh"
"$SCRIPT_DIR/mssql-wait.sh"

echo "Applying bootstrap SQL to TestDB"
docker exec -i "$CONTAINER_NAME" "$SQLCMD" \
  -S localhost \
  -U sa \
  -P "$MSSQL_SA_PASSWORD" \
  -C \
  -b \
  -i /dev/stdin < "$SCRIPT_DIR/mssql-bootstrap.sql"

echo "MSSQL test environment is ready."
