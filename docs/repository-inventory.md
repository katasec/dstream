# Repository Inventory

This document inventories the companion repositories in the current DStream workspace and summarizes their role in the overall ecosystem.

## Companion Repositories

| Repository | Origin | Primary Role | Purpose | Tech Stack | Integration with DStream CLI | Status |
|---|---|---|---|---|---|---|
| `dstream-ingester-mssql` | Extracted from core `dstream` | Input provider | Reads SQL Server CDC changes and emits change events as JSON envelopes over stdout. Includes checkpointing, distributed locking, and recovery behaviors designed to preserve at-most-once delivery characteristics under normal operation. Details: [docs/plugins/mssql-ingester.md](plugins/mssql-ingester.md). | Go | Referenced as an input provider via `provider_ref` or local `provider_path` in HCL tasks. | Mature |
| `dstream-log-output-provider` | Net-new provider repo | Output provider | Receives stream events and logs them to stderr, including graceful handling of non-JSON lines for troubleshooting and debugging pipelines. | Go | Referenced as an output provider in HCL; useful for diagnostics and development runs. | Active |
| `dstream-console-output-provider` | Net-new provider repo | Output provider | Displays streamed events in human-readable formats (`simple`, `structured`, `json`) for local inspection and demos. | .NET (C#) | Referenced as an output provider in HCL, often paired with test input providers. | Active |
| `dstream-dotnet-sdk` | Ecosystem SDK repo | Provider SDK | SDK for building custom DStream input/output providers with minimal stdin/stdout plumbing work. Defines abstractions and host runtime helpers for provider authors. | .NET (C#) | Used by .NET-based providers; enables third-party plugin development for the DStream ecosystem. | Active |
| `dstream-counter-input-provider` | Net-new provider repo | Input provider | Generates synthetic counter events at configurable intervals for testing, demos, and performance checks. | .NET (C#) | Referenced as an input provider in HCL for quick-start pipelines and baseline testing. | Active |

Status legend: `Mature` = established implementation with significant production-like history, `Active` = currently usable in common DStream flows, `In progress` = functional foundation exists but still being hardened.

## Ecosystem Shape

- DStream CLI is the orchestrator and router in the middle.
- Input providers generate event streams to stdout.
- Output providers consume those event streams from stdin.
- Providers are distributed independently (typically via OCI references), enabling multi-repo and multi-language plugin development.
