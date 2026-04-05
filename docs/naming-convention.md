# DStream Naming Convention

> Established 2026-04-05. Apply to all new repos; rename legacy repos when convenient.

## Pattern

```
dstream-{classifier}-{technology}
```

Three segments. The classifier is the second word and tells you what the repo is.

## Classifiers

| Classifier | Meaning |
|------------|---------|
| `in` | Input provider — reads data from a source |
| `out` | Output provider — writes data to a destination |
| `sdk` | SDK for building providers |

Extensible: `lib`, `tool`, etc. as needed.

## Current Ecosystem

| Repo | Role |
|------|------|
| `dstream` | Core CLI / orchestrator |
| `dstream-sdk-dotnet` | .NET provider SDK |
| `dstream-in-mssql` | MSSQL CDC input provider |
| `dstream-in-counter` | Counter input (test/demo) |
| `dstream-out-asb` | Azure Service Bus output provider |
| `dstream-out-console` | Console output (debug) |
| `dstream-out-log` | Log output (debug) |

## Legacy Names (To Rename)

| Current | Target |
|---------|--------|
| `dstream-ingester-mssql` | `dstream-in-mssql` |
| `dstream-dotnet-sdk` | `dstream-sdk-dotnet` |
| `dstream-counter-input-provider` | `dstream-in-counter` |
| `dstream-console-output-provider` | `dstream-out-console` |
| `dstream-log-output-provider` | `dstream-out-log` |
