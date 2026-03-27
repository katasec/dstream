#!/usr/bin/env bash
set -euo pipefail

CONTAINER_NAME="${CONTAINER_NAME:-sqlserver}"
MSSQL_SA_PASSWORD="${MSSQL_SA_PASSWORD:-Passw0rd123}"
SQLCMD="/opt/mssql-tools18/bin/sqlcmd"

run_scalar() {
  local query="$1"
  docker exec "$CONTAINER_NAME" "$SQLCMD" \
    -S localhost \
    -U sa \
    -P "$MSSQL_SA_PASSWORD" \
    -C \
    -d master \
    -h -1 \
    -W \
    -Q "SET NOCOUNT ON; $query" | awk 'NF { print; exit }' | tr -d '[:space:]'
}

assert_eq() {
  local name="$1"
  local expected="$2"
  local actual="$3"
  if [[ "$actual" != "$expected" ]]; then
    echo "FAIL: $name (expected=$expected, actual=$actual)" >&2
    exit 1
  fi
  echo "PASS: $name"
}

echo "Verifying MSSQL CDC test prerequisites"

db_exists=$(run_scalar "SELECT COUNT(*) FROM sys.databases WHERE name='TestDB';")
assert_eq "TestDB exists" "1" "$db_exists"

db_cdc_enabled=$(run_scalar "SELECT CAST(is_cdc_enabled AS INT) FROM sys.databases WHERE name='TestDB';")
assert_eq "TestDB CDC enabled" "1" "$db_cdc_enabled"

persons_exists=$(run_scalar "SELECT COUNT(*) FROM TestDB.sys.tables WHERE name='Persons' AND schema_id=SCHEMA_ID('dbo');")
assert_eq "dbo.Persons exists" "1" "$persons_exists"

cars_exists=$(run_scalar "SELECT COUNT(*) FROM TestDB.sys.tables WHERE name='Cars' AND schema_id=SCHEMA_ID('dbo');")
assert_eq "dbo.Cars exists" "1" "$cars_exists"

persons_columns=$(run_scalar "SELECT COUNT(*) FROM TestDB.sys.columns WHERE object_id=OBJECT_ID('TestDB.dbo.Persons') AND name IN ('ID','Name','Description');")
assert_eq "Persons has ID/Name/Description" "3" "$persons_columns"

cars_columns=$(run_scalar "SELECT COUNT(*) FROM TestDB.sys.columns WHERE object_id=OBJECT_ID('TestDB.dbo.Cars') AND name IN ('ID','Name','Description');")
assert_eq "Cars has ID/Name/Description" "3" "$cars_columns"

persons_id_identity=$(run_scalar "SELECT CASE WHEN EXISTS(SELECT 1 FROM TestDB.sys.identity_columns WHERE object_id=OBJECT_ID('TestDB.dbo.Persons') AND name='ID') THEN 1 ELSE 0 END;")
assert_eq "Persons.ID is identity" "1" "$persons_id_identity"

cars_id_identity=$(run_scalar "SELECT CASE WHEN EXISTS(SELECT 1 FROM TestDB.sys.identity_columns WHERE object_id=OBJECT_ID('TestDB.dbo.Cars') AND name='ID') THEN 1 ELSE 0 END;")
assert_eq "Cars.ID is identity" "1" "$cars_id_identity"

persons_cdc=$(run_scalar "SELECT COUNT(*) FROM TestDB.cdc.change_tables WHERE source_object_id=OBJECT_ID('TestDB.dbo.Persons');")
assert_eq "Persons CDC enabled" "1" "$persons_cdc"

cars_cdc=$(run_scalar "SELECT COUNT(*) FROM TestDB.cdc.change_tables WHERE source_object_id=OBJECT_ID('TestDB.dbo.Cars');")
assert_eq "Cars CDC enabled" "1" "$cars_cdc"

echo "All MSSQL CDC prerequisites are satisfied."
