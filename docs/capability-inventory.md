# DStream Capability Inventory

> Comprehensive inventory of what exists today across the dstream ecosystem.
> Written 2026-04-05 during the Capability Reset workstream.

---

## 1. DStream CLI

**Repo**: `~/progs/dstream` | **Language**: Go | **Entry**: `main.go` -> `cmd/`

### Commands

| Command | Description |
|---------|-------------|
| `run <task>` | Execute a streaming pipeline |
| `init <task>` | Initialize output provider infrastructure |
| `plan <task>` | Preview infrastructure changes (Terraform-style) |
| `status <task>` | Show current infrastructure status |
| `destroy <task>` | Tear down infrastructure resources |

Global flags: `--config/-c` (HCL file, default `dstream.hcl`), `--log-level/-l`, `--log-format/-f`, `--log-time/-t`

### Execution Modes

**Provider mode** (`type = "providers"`) — Primary, current architecture.
Three-process orchestration:

```
[Input Provider] --stdout--> [DStream CLI relay] --stdin--> [Output Provider]
```

- Resolves provider binaries via `provider_path` (local) or `provider_ref` (OCI/ORAS pull)
- Sends command envelope to each provider's stdin on startup
- Relays JSON lines from input stdout to output stdin (transparent, line-by-line)
- Forwards both providers' stderr to CLI stderr
- Input providers always receive `"command": "run"` regardless of CLI command
- Output providers receive the actual lifecycle command (run/init/plan/status/destroy)
- Graceful shutdown: SIGTERM -> 10s grace -> SIGKILL
- 5-minute execution timeout per task

**Legacy plugin mode** (`type = "plugin"`) — gRPC via HashiCorp go-plugin. Still functional, not primary. Uses protobuf service definition in `proto/plugin.proto`. Supports only `run` (no lifecycle commands).

### HCL Configuration

```hcl
task "my-task" {
  type = "providers"

  input {
    provider_ref  = "ghcr.io/katasec/dstream-ingester-mssql:v0.0.55"
    # or provider_path = "../local/binary"
    config {
      # arbitrary provider-specific fields
    }
  }

  output {
    provider_ref  = "ghcr.io/katasec/some-output-provider:v0.1.0"
    config {
      # arbitrary provider-specific fields
    }
  }
}
```

- Parsed with HashiCorp HCL v2 + gohcl
- Supports Sprig template functions (`{{ env "VAR" }}`, `{{ date "..." }}`, etc.)
- Config blocks are late-bound (raw HCL body) — no schema whitelist in CLI
- Type-preserving conversion to JSON (strings, numbers, bools, lists, objects)

### Protocol: Command Envelope

Sent once to provider stdin on startup:

```json
{
  "command": "run|init|plan|status|destroy",
  "config": { ... provider-specific fields ... }
}
```

### Protocol: Data Envelope

JSON lines (one per line) flowing input stdout -> CLI relay -> output stdin:

```json
{
  "metadata": { "table": "users", "operation": "insert", "sequence": 42 },
  "data": { "id": 123, "name": "John Doe" }
}
```

### OCI/ORAS Provider Resolution

- Cache: `~/.dstream/plugins/<name>/<version>/<platform>/plugin`
- Platform detection: `runtime.GOOS_runtime.GOARCH`
- Pulls via `oras pull` CLI, renames to `plugin`, `chmod 0755`
- Cache-first: skips pull if binary already cached
- Requires `~/.oras-config` for registry auth

### Embedded Code (Unused by Provider Mode)

The following exist in `internal/publisher/` but are **not called** from provider-mode execution:

- Azure Service Bus publisher (`internal/publisher/messaging/azure/servicebus/`)
- Azure Event Hub publisher (`internal/publisher/messaging/azure/eventhub/`)
- Console publisher (`internal/publisher/debug/console/`)
- Publisher factory with planned types: s3, mongodb, sql

These are remnants of the single-binary era. All output is now delegated to external providers.

### Legacy gRPC Code (Functional but Deprecated)

- `proto/plugin.proto` — service definition (GetSchema, Start)
- `pkg/plugins/` — client/server wrappers, serve function
- Handshake: protocol v1, magic cookie `DSTREAM_PLUGIN`
- Used when task `type = "plugin"` or blank

---

## 2. MSSQL CDC Ingester

**Repo**: `~/progs/dstream-ingester-mssql` | **Language**: Go | **gRPC status**: Fully migrated to stdin/stdout

### Core Capabilities

| Feature | Description |
|---------|-------------|
| CDC reading | Queries `cdc.dbo_<table>_CT` using `sys.fn_cdc_get_all_changes_*` pattern |
| Multi-table monitoring | One goroutine per table, shared DB connection, independent tracking |
| Checkpointing | Per-table LSN + seqval persistence in `cdc_offsets` table (auto-created) |
| Deduplication | Dual checkpoint (LSN + seqval) prevents both cross-transaction and within-transaction duplicates |
| Distributed locking | Azure Blob Storage lease-based locks, 2-min TTL with stale breaking |
| Exponential backoff | Configurable initial/max intervals, doubles on idle, resets on changes |
| Dynamic batch sizing | Calculates optimal batch from avg row size vs SKU limits (256KB std / 1MB premium), resamples hourly |
| Graceful shutdown | SIGINT/SIGTERM -> context cancellation -> all monitors stop cleanly |

### Communication Protocol

**Input** (stdin, first line): Accepts both direct JSON config and DStream command envelope format.

```json
{
  "db_connection_string": "server=localhost;database=testdb;...",
  "poll_interval": "5s",
  "max_poll_interval": "5m",
  "tables": ["dbo.customers", "dbo.orders"],
  "lock_config": {
    "type": "azure_blob",
    "connection_string": "DefaultEndpointsProtocol=https;...",
    "container_name": "locks"
  }
}
```

**Output** (stdout, JSON lines):

```json
{
  "metadata": {
    "TableName": "dbo.customers",
    "LSN": "00000020000001000001",
    "Seq": "00000000000000000002",
    "OperationID": 2,
    "OperationType": "Insert"
  },
  "data": {
    "CustomerID": "123",
    "Name": "John Doe",
    "Email": "john@example.com"
  }
}
```

**Logging**: All to stderr with `[MSSQL-CDC]` prefix. Level configurable via `DSTREAM_LOG_LEVEL`.

### Checkpoint Strategy

- Storage: `cdc_offsets` table in the monitored database (auto-created, auto-migrated)
- Columns: `table_name` (PK), `last_lsn` (VARBINARY), `last_seq` (VARBINARY), `updated_at`
- UPSERT via `MERGE INTO` on save
- Recovery: on restart, resumes from last saved LSN+Seq; first run starts from beginning

### Locking Strategy

- Azure Blob Storage leases (native atomic lock primitive)
- Lock path: `server_name/table_name.lock`
- Infinite lease with 2-min stale detection
- On conflict: checks Last Modified, breaks stale leases
- Table skipped if locked by another instance (graceful, retries next cycle)
- Supports `"type": "none"` for local/testing scenarios

### Failure Recovery

- DB connection errors: backoff + retry, monitor continues
- CDC query errors: logged, checkpoint NOT updated, backoff + retry
- Lock failures: table skipped this cycle, retried next
- Publish failures: batch retried (checkpoint not updated)
- Per-table isolation: one table failing doesn't affect others

### Build & Packaging

- Cross-platform: linux/darwin amd64+arm64, windows amd64
- Binary names: `plugin.<platform>`
- Stripped: `-ldflags="-s -w"`, `CGO_ENABLED=0`
- Published to GHCR via ORAS (`make push`)

---

## 3. .NET Provider SDK

**Repo**: `~/progs/dstream-dotnet-sdk` | **Language**: C# (.NET 9) | **Version**: 0.1.1

### SDK Packages

| Package | Purpose |
|---------|---------|
| `Katasec.DStream.Abstractions` | Pure interfaces, no dependencies |
| `Katasec.DStream.SDK.Core` | Runtime: base classes, StdioProviderHost, protocol handling |

### Provider Abstractions

```
IProvider (marker)
  IInputProvider    -> ReadAsync() returns IAsyncEnumerable<Envelope>
  IOutputProvider   -> WriteAsync() receives IEnumerable<Envelope> batch
  IInfrastructureProvider -> InitializeAsync, DestroyAsync, GetStatusAsync, PlanAsync
```

**Base classes**: `ProviderBase<TConfig>` (config injection, context, logger) and `InfrastructureProviderBase<TConfig>` (adds lifecycle template methods).

**Data model**: `Envelope(object Payload, IReadOnlyDictionary<string, object?> Meta)`

### StdioProviderHost

Handles all protocol plumbing for provider authors:

1. Reads JSON config from stdin (line 1)
2. Parses command envelope or direct config
3. Instantiates provider, injects config
4. Routes to appropriate handler (run/init/plan/status/destroy)
5. For input: calls ReadAsync, serializes envelopes to stdout
6. For output: reads envelopes from stdin, deserializes, calls WriteAsync
7. For infra: calls lifecycle method, serializes InfrastructureResult to stdout

**Entry point** (5 lines):

```csharp
await StdioProviderHost.RunProviderWithCommandAsync<MyProvider, MyConfig>();
```

### Infrastructure Lifecycle

Output providers can optionally implement `IInfrastructureProvider` to manage their destination resources. The CLI routes `init/plan/status/destroy` commands to these methods. Returns `InfrastructureResult` with status, resources, metadata, message, error.

### Sample Providers

| Sample | Type | Purpose |
|--------|------|---------|
| `counter-input-provider` | Input | Generates sequential counter events at configurable intervals |
| `console-output-provider` | Output + Infra | Formats and displays events; demonstrates infrastructure lifecycle |

### Tests

Placeholder infrastructure only — `TestKit` and `Providers.AsbQueue.Tests` exist but are empty. Pre-stable SDK, tests deferred until API solidifies.

### Publishing

- NuGet packages via GitHub Actions on version tags
- Version in `VERSION.txt`, bumped via `scripts/version-bump.ps1`
- Self-contained single-file binaries for sample providers (~68MB with runtime)

---

## 4. Ecosystem Providers (External Repos)

| Provider | Type | Stack | Status |
|----------|------|-------|--------|
| `dstream-log-output-provider` | Output | Go | Active |
| `dstream-console-output-provider` | Output | .NET | Active |
| `dstream-counter-input-provider` | Input | .NET | Active |

---

## 5. Gap Analysis for Capability Reset

### What we had (single binary era)
MSSQL CDC -> embedded dstream -> Azure Service Bus queues. One binary, gRPC internal comms. Production-tested, scaled well.

### What we have now
- MSSQL ingester: extracted, fully migrated to stdin/stdout, feature-complete
- DStream CLI: provider-mode orchestration working, relay pipeline functional
- SDK: abstractions and host runtime ready for building output providers
- Azure Service Bus output provider: **does not exist yet**

### What's needed for Reset

1. **Verify ingester E2E with CLI** — confirm the MSSQL ingester works through the CLI relay (not just standalone). Test with `dstream run` against a live SQL Server.
2. **Build Azure Service Bus output provider** — using the .NET SDK. Needs: WriteAsync to send messages to queues/topics, infrastructure lifecycle to create/manage queues.
3. **E2E validation** — MSSQL CDC -> dstream CLI -> Azure Service Bus. Match original capability.
4. **Clean up dead code** — embedded publishers in `internal/publisher/` and legacy gRPC path (decision needed on timing).
