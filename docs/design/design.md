# System Design

> Design rationale and architectural decisions. Update when core architecture, data models, or system design changes.
> See the [docs freshness check](../../AGENTS.md) in AGENTS.md.

## Design Goals

1. **Terraform-style UX for streaming**: Most users should compose pipelines in HCL and reuse providers without writing custom connector code.
2. **Plugin ecosystem over monolith**: DStream should orchestrate independently versioned providers, not bundle all integrations into one binary.
3. **Language and platform agnostic providers**: Provider development must work with any language/runtime that can read/write JSON over stdin/stdout.
4. **Low operational overhead**: Keep runtime and plugin contracts simple enough for DevOps teams to operate without deep data platform specialization.
5. **Reproducible distribution**: Provider artifacts should be versioned, immutable, and fetchable from OCI registries.

## Product Thesis

DStream is built around a Data DevOps thesis: data movement and streaming orchestration should be operable by infrastructure/platform engineers using declarative configuration, without requiring every team to maintain custom integration code or specialized data engineering workflows.

Core framing:

1. Most users compose tasks from existing providers.
2. A smaller ecosystem segment builds providers.
3. Standard contracts + independent provider repos create ecosystem leverage.

## Origin Story and Initial Beachhead

Initial use cases were driven by SQL Server CDC requirements from consulting delivery work where frequent managed orchestration jobs (for example, high-frequency Azure Data Factory patterns) introduced avoidable cost and operational overhead.

The first practical wedge was:

1. Capture incremental SQL changes via CDC.
2. Stream them through composable provider pipelines.
3. Reduce cost/complexity compared with tightly coupled source-destination jobs.

This origin explains current design priorities: checkpointing, stream reliability boundaries, and low-friction provider composition.

## Architecture Decisions

### AD-1: Standardize provider runtime on stdin/stdout JSON lines

**Context**: Early designs supported a gRPC/go-plugin execution model. This increased implementation overhead and language/runtime coupling for external provider authors.
**Decision**: Use a simple process contract: command envelope + JSON line stream over stdin/stdout.
**Rationale**: This is the lowest-friction interface for cross-language provider development (including shell-based tooling) and aligns with Unix pipeline principles.
**Consequences**: Runtime protocol remains intentionally minimal, but contract semantics must be documented clearly to avoid provider drift.

### AD-2: Use a three-process orchestration model

**Context**: DStream needs a clean separation between source and destination connectors while preserving process isolation.
**Decision**: Execute one input provider process and one output provider process, with DStream CLI relaying stream messages.
**Rationale**: Clear ownership boundaries, easier debugging, and constrained blast radius when one provider fails.
**Consequences**: DStream must manage process lifecycle, backpressure behavior, relay errors, and shutdown coordination.

### AD-3: Distribute providers as OCI artifacts via ORAS

**Context**: Providers live in independent repositories and should be shipped without coupling to DStream release cadence.
**Decision**: Resolve `provider_ref` to OCI artifact references and pull binaries with ORAS into a local cache.
**Rationale**: Reuses existing registry infrastructure, supports semantic versioning, and mirrors Terraform-like provider distribution expectations.
**Consequences**: Runtime depends on ORAS availability and artifact naming conventions; artifact compatibility and signing policy become ecosystem concerns.

### AD-4: Keep HCL as declarative control plane

**Context**: Users need a familiar way to declare source, destination, and provider-specific configuration.
**Decision**: Task definitions are HCL-based, with `type = "providers"` as the primary execution mode.
**Rationale**: Declarative workflow lowers adoption cost for infrastructure-minded users and keeps pipeline definitions version-control friendly.
**Consequences**: Schema/documentation quality is critical, and parser/runtime parity must be maintained as protocol evolves.

### AD-5: Favor small, composable tasks over monolithic pipelines

**Context**: Real-world workflows often require staged transformation, buffering, and delivery concerns.
**Decision**: Encourage each task to do one thing well (ingest/normalize, route/buffer, deliver), then compose tasks through durable intermediates.
**Rationale**: Smaller tasks are easier to reason about, test, scale, and operate; durable intermediates isolate failures and support replay/recovery.
**Consequences**: Requires clear inter-task event contracts and guidance on where transformation responsibility lives.

## Runtime Flow (Current State)

1. User runs `dstream run <task-name>`.
2. DStream loads task configuration from HCL and resolves task type.
3. For provider tasks, DStream resolves input/output binaries via `provider_path` or `provider_ref`.
4. If `provider_ref` is used, DStream pulls artifact via ORAS and reuses local cache when present.
5. DStream starts input and output processes.
6. DStream sends one command envelope JSON payload to each provider stdin.
7. Input provider emits data envelopes as JSON lines on stdout.
8. DStream relays each input line to output provider stdin.
9. DStream forwards provider stderr for logs and coordinates graceful shutdown.

## Composable Task Pattern

DStream tasks are intentionally simple building blocks. A common production pattern is chaining two tasks via a queue/bus boundary.

Example pattern:

1. Task A: Source-specific extraction and normalization.
2. Task A output: Durable buffer (for example Azure Service Bus).
3. Task B: Subscription/consumer input.
4. Task B output: Destination-specific delivery (for example Twilio, database, webhook).

Benefits:

1. Independent scaling and retry boundaries.
2. Easier incident isolation and replay.
3. Reusable destination delivery task across multiple sources.
4. Clearer security boundaries for destination credentials.

## Provider Protocol (Current State)

- **Config/control message**: First message on stdin is a command envelope:
	- `{"command":"run|init|plan|status|destroy","config":{...}}`
- **Input providers**:
	- Receive command envelope.
	- Emit stream events to stdout as JSON lines.
- **Output providers**:
	- Receive command envelope.
	- Consume JSON lines from stdin and write to destination system.
- **Logging**:
	- Providers should write logs to stderr, not stdout.

Note: In non-`run` lifecycle commands, the current orchestrator sends `run` to input providers and the requested lifecycle command to output providers.

## Data Model

Core HCL entities:

- `task`:
	- `name` (label)
	- `type` (primary mode: `providers`)
	- legacy plugin fields (`plugin_path`, `plugin_ref`)
- `input`:
	- `provider_path` or `provider_ref`
	- provider-specific `config` block
- `output`:
	- `provider_path` or `provider_ref`
	- provider-specific `config` block

Runtime payload shapes:

- Command envelope sent by DStream to providers.
- Data envelopes emitted by input providers and forwarded unchanged to output providers.

## System Boundaries

### DStream Owns

- HCL task parsing and task selection.
- Provider binary resolution and local caching integration.
- Process orchestration (start, relay, shutdown, signal handling).
- Command routing and stdin/stdout transport contract.

### Providers Own

- Source-specific ingestion logic (input providers).
- Destination-specific delivery logic (output providers).
- Provider-level retries, serialization details, destination semantics.

### External Dependencies Own

- OCI registry availability and artifact hosting.
- ORAS CLI execution and artifact pull behavior.
- Destination system guarantees (for example, queue semantics, DB consistency).

## Current Risks and Gaps

1. **Dual architecture surface**: Legacy gRPC/go-plugin path still exists alongside provider mode, increasing cognitive overhead.
2. **Protocol formalization gap**: No explicit protocol version negotiation in command/data envelope format.
3. **Lifecycle asymmetry**: Non-`run` commands currently target output provider only, which may surprise plugin authors.
4. **Documentation drift risk**: README positioning and implementation details can diverge without a single canonical protocol spec.
5. **Ecosystem readiness gap**: Provider author guidance exists, but stronger compatibility tests and packaging conventions are needed for broad OSS adoption.
6. **Historical backfill gap**: CDC streaming baseline is present, but snapshot/bootstrap transfer of historical data is not yet designed as a first-class capability.