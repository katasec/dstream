#!/usr/bin/env bash
set -euo pipefail

CONTAINER_NAME="${CONTAINER_NAME:-sqlserver}"

if docker ps -a --format '{{.Names}}' | grep -Fxq "$CONTAINER_NAME"; then
  echo "Stopping and removing container: $CONTAINER_NAME"
  docker rm -f "$CONTAINER_NAME" >/dev/null
else
  echo "Container not found: $CONTAINER_NAME"
fi

if docker ps -a --format '{{.Names}}' | grep -Fxq "$CONTAINER_NAME"; then
  echo "Failed to remove container: $CONTAINER_NAME" >&2
  exit 1
fi

echo "MSSQL container is stopped and removed."
