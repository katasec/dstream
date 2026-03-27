#!/usr/bin/env bash
set -euo pipefail

CONTAINER_NAME="${CONTAINER_NAME:-sqlserver}"
MSSQL_IMAGE="${MSSQL_IMAGE:-mcr.microsoft.com/mssql/server:2019-latest}"
MSSQL_SA_PASSWORD="${MSSQL_SA_PASSWORD:-Passw0rd123}"
HOST_PORT="${HOST_PORT:-1433}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONF_PATH="${CONF_PATH:-$SCRIPT_DIR/mssql.conf}"

if [[ ! -f "$CONF_PATH" ]]; then
  echo "Missing config file: $CONF_PATH" >&2
  exit 1
fi

echo "Starting SQL Server container: $CONTAINER_NAME"

if docker ps -a --format '{{.Names}}' | grep -Fxq "$CONTAINER_NAME"; then
  docker rm -f "$CONTAINER_NAME" >/dev/null
fi

docker run --platform linux/amd64 \
  -e "ACCEPT_EULA=Y" \
  -e "MSSQL_SA_PASSWORD=$MSSQL_SA_PASSWORD" \
  -p "$HOST_PORT:1433" \
  --name "$CONTAINER_NAME" \
  -v "$CONF_PATH:/var/opt/mssql/mssql.conf" \
  -d "$MSSQL_IMAGE" >/dev/null

echo "SQL Server container started."
