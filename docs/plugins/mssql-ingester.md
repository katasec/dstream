# MSSQL Ingester Plugin

This page documents the DStream SQL Server CDC input plugin repository (`dstream-ingester-mssql`) and its current behavior.

## Purpose

The MSSQL ingester is an input provider that reads SQL Server CDC change rows and emits DStream-compatible JSON envelopes to stdout.

It is intended to be used in provider-mode tasks where DStream CLI orchestrates:

1. Input provider process (this plugin)
2. DStream relay process
3. Output provider process

## Repository

- Name: `dstream-ingester-mssql`
- Origin: Extracted from the original core DStream codebase into a separate provider repository
- Runtime: Go binary, stdin/stdout protocol

## Configuration

The plugin reads a command envelope from stdin. The `config` object is expected to contain the provider settings below.

### Config fields

- `db_connection_string`: SQL Server connection string
- `poll_interval`: Base polling interval (for example `5s`)
- `max_poll_interval`: Maximum polling backoff interval (for example `30s`, `5m`)
- `tables`: List of source table names to monitor (for example `Persons`, `Cars`)
- `lock_config`:
  - `type`: Lock provider type (`azure_blob` or `none` depending on deployment expectations)
  - `connection_string`: Lock backend connection string (when applicable)
  - `container_name`: Blob container name (when applicable)

### Example HCL usage

```hcl
task "mssql-test" {
  type = "providers"

  input {
    provider_ref = "ghcr.io/katasec/dstream-ingester-mssql:v0.0.55"
    config {
      db_connection_string = "server=localhost,1433;user id=sa;password=Passw0rd123;database=TestDB;encrypt=disable"
      poll_interval = "5s"
      max_poll_interval = "30s"
      tables = ["Persons", "Cars"]
      lock_config = {
        type = "none"
      }
    }
  }

  output {
    provider_ref = "ghcr.io/katasec/dstream-log-output-provider:v0.1.0"
    config {
      logLevel = "info"
    }
  }
}
```

## How it works

For each configured table, the plugin:

1. Initializes checkpoint persistence table (`cdc_offsets`) if needed.
2. Loads last `(LSN, Seq)` checkpoint for that table.
3. Polls CDC table incrementally using ordered `(start_lsn, seqval)` queries.
4. Publishes a batch of detected changes.
5. Only after publish success, advances in-memory cursor and persists new checkpoint.
6. Uses exponential backoff when no changes are detected.

## Delivery semantics (current behavior)

### What is strong today

- Incremental CDC cursoring with `(LSN, Seq)` checkpoint avoids re-reading already committed cursor positions in normal operation.
- Checkpoint is not advanced on publish failure, preventing data loss from failed sends.

### Important boundary conditions

- A crash after successful publish but before checkpoint persistence can cause replay on restart.
- Distributed lock implementations exist in the codebase, but active runtime wiring should be treated as implementation-dependent and verified in integration environments.

Practical interpretation: current behavior is generally replay-safe and duplicate-minimizing, but strict global exactly-once semantics should not be assumed without downstream idempotency and full lock-path validation.

## Operational notes

- The plugin emits data envelopes to stdout; logs should go to stderr.
- Poll intervals and max backoff should be tuned for source change rate and load.
- Downstream consumers should be prepared for occasional duplicate replay under failure/restart windows.

## Related docs

- DStream runtime architecture: [docs/design/design.md](../design/design.md)
- Repository inventory: [docs/repository-inventory.md](../repository-inventory.md)
