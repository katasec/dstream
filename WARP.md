# WARP.md - DStream Project Master Context

**Last Updated**: October 23, 2025  
**Current Priority**: Create Log Output Provider  
**Project Status**: Phases 0-2 Complete ✅ | Phase 3 Ready (HCL Locals)

---

## 🚀 Quick Start for New Sessions

This is your master context file for the DStream project. Read it to understand the current state.

**TL;DR - Current Status:**
- ✅ **Foundation Complete**: CLI infrastructure, .NET SDK, external providers all working
- ✅ **MSSQL CDC Provider**: Production-ready with real CDC queries (Go)
- ✅ **Modern Architecture**: stdin/stdout JSON communication
- 🎯 **Next Priority**: HCL `locals` support to eliminate table duplication

---

## 📊 Project Status Overview

| Phase | Status | Description |
|-------|--------|-------------|
| **Phase 0** | ✅ COMPLETE | Foundation & CLI Infrastructure |
| **Phase 1** | ✅ COMPLETE | .NET SDK & Provider Architecture |
| **Phase 2** | ✅ COMPLETE | External Provider Pattern & OCI Distribution |
| **Phase 3** | 🎯 NEXT | HCL Locals Support (Single source of truth) |

---

## 🎯 **IMMEDIATE PRIORITY: Create Log Output Provider**

### Problem: Console Output Provider Tries to Parse Logs as JSON

When running `mssql-test` task, the console output provider receives log lines from the MSSQL provider and tries to parse them as JSON:

```
Failed to parse JSON: 2025/10/23 21:23:04 [INFO] Loaded configuration for 2 tables
```

### Solution: Create a Simple Log Output Provider

A new Go output provider that:
1. Reads command envelope from stdin (first line)
2. For each subsequent line, parses as JSON if possible, logs as-is if not
3. Writes everything to stderr with structured logging
4. No JSON parsing failures - just logs data

**Repository**: `~/progs/dstream/dstream-log-output-provider/` (new repo)

**Structure**:
```
dstream-log-output-provider/
├── main.go                    ← Simple log provider
├── go.mod
├── Makefile                   ← Cross-platform build
├── scripts/
│   ├── build.sh              ← Build binaries
│   ├── push.sh               ← Push OCI image
│   └── create-manifest.sh    ← Provider manifest
├── provider.json             ← Metadata
└── README.md                 ← Documentation
```

**Quick Start**:
```bash
# 1. Create the repository
cd ~/progs/dstream
mkdir dstream-log-output-provider
cd dstream-log-output-provider

# 2. Initialize Go project
go mod init github.com/katasec/dstream-log-output-provider
echo 'package main' > main.go
# (Add provider code)

# 3. Build and test locally
go build -o log-output-provider
echo '{"command":"run","config":{}}' | ./log-output-provider
echo '{"data": {"id": 1}}' | ./log-output-provider  # Should log, not error

# 4. Create build/push scripts (similar to MSSQL provider)
# 5. Create Makefile
# 6. Build cross-platform: make build
# 7. Push to GHCR: make push
```

**After Creation**:
- Update `dstream.hcl` to use log provider:
  ```hcl
  output {
    provider_ref = "ghcr.io/katasec/dstream-log-output-provider:v0.1.0"
    config {
      logLevel = "info"
    }
  }
  ```
- Test: `go run . run mssql-test` (no more JSON parse errors)

**Key Implementation Details**:
- Read command envelope: `{"command":"run","config":{...}}`
- Read data lines and attempt JSON parse (for pretty printing)
- If JSON parse fails, log as-is to stderr
- Output nothing to stdout (stderr only for logging)
- Handle graceful shutdown

---

## 🎯 **NEXT PRIORITY: HCL Locals Support**

### Problem: Table Duplication Risk

Current config duplicates table lists:
```hcl
# ❌ CURRENT - Tables repeated in two places
task "mssql-cdc-to-asb" {
  input {
    config {
      tables = ["Orders", "Customers", "Products"]  # List 1
    }
  }
  output {
    config {
      tables = ["Orders", "Customers", "Products"]  # List 2 - can drift!
    }
  }
}
```

### Solution: HCL Locals Block

```hcl
# ✅ SOLUTION - Single source of truth
locals {
  tables = ["Orders", "Customers", "Products"]
}

task "mssql-cdc-to-asb" {
  input {
    config {
      tables = local.tables  # Single reference
    }
  }
  output {
    config {
      tables = local.tables  # Same reference - no duplication!
    }
  }
}
```

### Implementation: 2-3 hours

**Phase 1 (30-45 min): Discovery**
- [ ] Find HCL parsing: `pkg/config/config_funcs.go` - `DecodeHCL[T any]()`
- [ ] Check template order: `RenderHCLTemplate()` → `DecodeHCL()`
- [ ] Current code: `gohcl.DecodeBody(f.Body, nil, &c)` with nil context
- [ ] Target struct: `RootHCL` in `pkg/config/config.go`

**Phase 2 (30-45 min): Implementation**
- [ ] Add `Locals map[string]interface{}` to `RootHCL` struct
- [ ] Replace `gohcl.DecodeBody(f.Body, nil, &c)` with `hcl.EvalContext` containing locals
- [ ] Keep backward compatible (existing configs work unchanged)

**Phase 3 (30-45 min): Testing**
- [ ] Test no locals block (backward compatibility)
- [ ] Test simple locals: `local.tables = [...]`
- [ ] Test templates in locals: `local.db = "{{ env \"DB_NAME\" }}"`
- [ ] Test error handling for undefined references

---

## 📁 Project Structure

```
~/progs/dstream/
├── WARP.md                                    ← You are here
├── readme.md                                  ← User documentation
├── dstream/                                   ← Go CLI
│   ├── main.go
│   ├── dstream.hcl                           ← Task configuration
│   ├── pkg/config/
│   │   ├── config_funcs.go                  ← TODO: Add locals support
│   │   └── config.go                        ← Add Locals field
│   └── cmd/                                  ← CLI commands
│
├── dstream-ingester-mssql/                   ← Go MSSQL CDC Provider ✅
│   ├── internal/cdc/sqlserver/monitor.go   ← Real CDC queries ✅
│   ├── main.go                               ← stdin/stdout interface
│   └── README.md
│
├── dstream-dotnet-sdk/                       ← .NET SDK (NuGet v0.1.1)
│   ├── sdk/
│   │   ├── Katasec.DStream.Abstractions/
│   │   └── Katasec.DStream.SDK.Core/
│   └── samples/
│
├── dstream-counter-input-provider/           ← External .NET Provider ✅
│   ├── Makefile                              ← make build/test/clean
│   └── Program.cs
│
├── dstream-console-output-provider/          ← External .NET Provider ✅
│   ├── Makefile
│   └── Writer.cs + Infrastructure.cs
│
└── dstream-log-output-provider/              ← Simple Go Output Provider (NEW)
    ├── main.go                               ← Logs all data to stderr
    ├── Makefile
    ├── scripts/
    │   ├── build.sh                         ← Cross-platform build
    │   └── push.sh                          ← Push to GHCR
    └── README.md
```

---

## 💻 Development Commands

### Build All

```bash
# Go CLI
cd ~/progs/dstream/dstream && go build -o dstream

# MSSQL CDC Provider (already works)
cd ~/progs/dstream/dstream-ingester-mssql && go build -o dstream-ingester-mssql

# .NET SDK
cd ~/progs/dstream/dstream-dotnet-sdk && /usr/local/share/dotnet/dotnet build dstream-dotnet-sdk.sln

# External providers
cd ~/progs/dstream/dstream-counter-input-provider && make build
cd ~/progs/dstream/dstream-console-output-provider && make build
```

### Test

```bash
# Full pipeline
cd ~/progs/dstream/dstream && go run . run counter-to-console

# Individual providers
cd ~/progs/dstream/dstream-counter-input-provider && make test
cd ~/progs/dstream/dstream-console-output-provider && make test

# CLI commands
cd ~/progs/dstream/dstream
go run . init counter-to-console
go run . plan counter-to-console
go run . status counter-to-console
go run . destroy counter-to-console
go run . run counter-to-console
```

---

## ✅ What's Working Now

**CLI Infrastructure**
- ✅ All lifecycle commands: `init`, `destroy`, `plan`, `status`, `run`
- ✅ HCL configuration parsing
- ✅ Process orchestration

**MSSQL CDC Provider** (`~/progs/dstream/dstream-ingester-mssql/`)
- ✅ Compiles successfully
- ✅ Real CDC queries: `sys.fn_cdc_get_all_changes_*`
- ✅ LSN/Seq checkpoint tracking
- ✅ Exponential backoff polling
- ✅ Concurrent multi-table monitoring
- ✅ Distributed locking (Azure Blob)

**Modern Provider Architecture** (stdin/stdout)
- ✅ Counter input provider
- ✅ Console output provider
- ✅ Infrastructure lifecycle support
- ✅ Command routing via JSON envelopes
- ✅ OCI distribution ready

---

## 🏗️ Architecture: Why stdin/stdout?

Switched from gRPC to **Unix stdin/stdout pipes** for:
- ✅ **10-50x faster** IPC (5μs vs 100-200μs)
- ✅ **Universal language support** (every language has stdin/stdout)
- ✅ **50+ years** of battle testing
- ✅ **Simple testing** (echo '{}' | ./provider)

---

## 📋 Data Format

All data flows as JSON envelopes:

```json
{
  "data": {
    "id": 123,
    "name": "John Doe"
  },
  "metadata": {
    "TableName": "dbo.persons",
    "OperationType": "Insert",
    "LSN": "0000004c000028200003"
  }
}
```

---

## 🚀 Next Steps

### Phase 1: Discovery (30-45 min) ← START HERE
1. Find HCL parsing: `pkg/config/config_funcs.go` - `DecodeHCL[T any]()`
2. Find config struct: `pkg/config/config.go` - `RootHCL`
3. Understand flow: `RenderHCLTemplate()` → `DecodeHCL()`
4. Find the line: `gohcl.DecodeBody(f.Body, nil, &c)` (nil context is key)

### Phase 2: Implementation (30-45 min)
1. Add `Locals map[string]interface{}` field to `RootHCL` struct
2. Replace `gohcl.DecodeBody(f.Body, nil, &c)` with `hcl.EvalContext` containing locals
3. Keep backward compatible (existing configs work unchanged)

### Phase 3: Testing (30-45 min)
1. Test backward compatibility: `go run . run counter-to-console`
2. Test locals: Create config with `locals { tables = [...] }`
3. Test templates in locals: `local.db = "{{ env \"DB_NAME\" }}"`

---

## ✅ Verification Checklist (Do This First!)

Before implementing, verify everything works:

```bash
# Test current pipeline
cd ~/progs/dstream/dstream
go run . run counter-to-console

# Build CLI
go build -o dstream

# Test providers
cd ~/progs/dstream/dstream-counter-input-provider && make test
cd ~/progs/dstream/dstream-console-output-provider && make test
```

All should pass ✅

---

## 🎯 Success Criteria

After implementation, **this config should work**:

```hcl
locals {
  tables = ["Orders", "Customers", "Products"]
}

task "mssql-cdc-to-asb" {
  input {
    config {
      tables = local.tables  # Single reference
    }
  }
  output {
    config {
      tables = local.tables  # Same reference - no duplication!
    }
  }
}
```

Running `go run . run mssql-cdc-to-asb` should:
- ✅ Parse locals block
- ✅ Resolve `local.tables` references
- ✅ Pass same table list to input and output providers
- ✅ No duplication or drift possible

---

## 📚 Key Files to Edit for HCL Locals

1. **`dstream/pkg/config/config.go`** - Add `Locals` field to `RootHCL` struct
2. **`dstream/pkg/config/config_funcs.go`** - Update `DecodeHCL()` to use `hcl.EvalContext` with locals

---

### This Week (After Locals)
1. Build Azure Service Bus output provider
2. Test MSSQL CDC → ASB end-to-end

### Next Sprint
3. Additional providers: PostgreSQL CDC, Kafka, etc.
4. Provider marketplace

---

## 🚫 Development Rules

### **NEVER use Dockerfile for OCI builds**
- ❌ **Don't create Dockerfile** for any providers
- ❌ **Don't use Docker** for containerization
- ✅ **Use alternative OCI build methods** (buildah, podman, etc.)
- ✅ **Focus on single-file binaries** for easy distribution
- ✅ **Cross-platform builds** instead of containers when possible

*Rationale*: Project avoids Docker dependency for build processes

---

**All old documentation files (DESIGN_NOTES.md, ROADMAP.md, etc.) have been consolidated into this single WARP.md file for clarity.**
