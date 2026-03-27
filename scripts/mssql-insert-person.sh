#!/usr/bin/env bash
set -euo pipefail

CONTAINER_NAME="${CONTAINER_NAME:-sqlserver}"
MSSQL_SA_PASSWORD="${MSSQL_SA_PASSWORD:-Passw0rd123}"
SQLCMD="/opt/mssql-tools18/bin/sqlcmd"

timestamp="$(date -u +%Y%m%d%H%M%S)"
random_suffix="$(printf '%04d' "$((RANDOM % 10000))")"
person_name="Person-${timestamp}-${random_suffix}"
person_description="Generated at ${timestamp}"

echo "Inserting a test row into TestDB.dbo.Persons"

docker exec "$CONTAINER_NAME" "$SQLCMD" \
  -S localhost \
  -U sa \
  -P "$MSSQL_SA_PASSWORD" \
  -C \
  -d master \
  -W \
  -Q "SET NOCOUNT ON; INSERT INTO TestDB.dbo.Persons (Name, Description) OUTPUT INSERTED.ID, INSERTED.Name, INSERTED.Description VALUES (N'${person_name}', N'${person_description}');"
