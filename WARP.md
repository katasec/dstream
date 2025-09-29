# WARP.md

> **💡 Starting a new Warp session?** See `WARP_CONTEXT_RESTORE.md` for copy-paste prompts to quickly restore full project context and continue development without losing time.

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## 🗺️ **Navigation - Key Documents**

- **`ROADMAP.md`** - 🎯 **Current development roadmap and priorities** (START HERE for planning)
- **`DESIGN_NOTES_PHASE_2_COMPLETE.md`** - ✅ Completed infrastructure lifecycle management 
- **`WARP_CONTEXT_RESTORE.md`** - 🚀 Quick context restoration for new Warp sessions

## ✅ **CURRENT STATUS (September 2024)**

**Foundation & Infrastructure Lifecycle: COMPLETE ✅**
- **NuGet Publishing**: Automated pipeline with v0.1.1 published ✅
- **External Provider Pattern**: Independent repos using NuGet packages ✅  
- **Infrastructure Commands**: All CLI commands (init/destroy/plan/status/run) working ✅
- **OCI Distribution**: Both provider_path and provider_ref working with GHCR ✅
- **Modern Architecture**: stdin/stdout providers with command routing ✅

**SQL Server CDC Provider: EXTRACTED & MODERNIZED ✅ (September 28, 2024)**
- **Legacy Code Extraction**: Successfully extracted from dstream 0.0.16 branch ✅
- **Modern Architecture**: Converted to stdin/stdout JSON interface ✅
- **Simplified Configuration**: Clean JSON structure with shared settings ✅
- **Concurrent Multi-Table**: Each table monitored independently ✅
- **Distributed Locking**: Azure Blob Storage coordination ✅
- **Comprehensive Documentation**: Complete README with all components ✅
- **Repository**: `github.com/katasec/dstream-ingester-mssql` (working code) ✅

**Next Priority**: Complete CDC implementation (actual SQL Server CDC queries)

### 📋 **SQL Server CDC Provider Modernization Details (September 28, 2024)**

**Problem Solved**: The legacy `dstream-ingester-mssql` repository contained old, non-working code using the deprecated gRPC plugin architecture. This was completely modernized to match current DStream patterns.

**Extraction Process**:
1. **Source Material**: Copied working components from dstream 0.0.16 branch
2. **Architecture Migration**: Converted from gRPC plugins to stdin/stdout JSON interface
3. **Configuration Simplification**: Eliminated repetitive per-table config in favor of shared settings
4. **Component Organization**: Proper internal package structure matching .NET provider patterns

**Key Components Extracted & Modernized**:
- **`internal/cdc/`**: Checkpoint management, backoff logic, batch sizing
- **`internal/locking/`**: Distributed locking with Azure Blob Storage
- **`internal/config/`**: JSON configuration parsing with simplified structure
- **`internal/db/`**: Database connection and metadata utilities
- **`pkg/types/`**: CDC event structures and interfaces
- **`main.go`**: Modern stdin/stdout entry point with concurrent table monitoring

**Configuration Improvement**:
```json
// OLD (Repetitive)
{"tables": [{"name": "dbo.orders", "db_connection_string": "...", "poll_interval": "5s", "lock_config": {...}}, ...]}

// NEW (Clean)
{"db_connection_string": "...", "poll_interval": "5s", "lock_config": {...}, "tables": ["dbo.orders", "dbo.customers"]}
```

**Repository State**:
- **Location**: `~/progs/dstream/dstream-ingester-mssql/` (local) + `github.com/katasec/dstream-ingester-mssql` (remote)
- **Status**: Compiles successfully, handles configuration, ready for CDC implementation
- **Documentation**: Comprehensive README covering all components and architecture
- **Next**: Implement actual CDC query logic using `sys.fn_cdc_get_all_changes_*`

## Development Commands

### Environment Configuration

**PowerShell on macOS:**
- .NET path: `/usr/local/share/dotnet/dotnet`
- Use full path when running dotnet commands in PowerShell
- ORAS binary: `/usr/local/bin/oras`
- Use full path when pushing OCI artifacts to container registries

### Building the Solution
```bash
# Navigate to the SDK directory
cd ~/progs/dstream/dstream-dotnet-sdk

# Build the entire solution
/usr/local/share/dotnet/dotnet build dstream-dotnet-sdk.sln

# Build a specific project
/usr/local/share/dotnet/dotnet build sdk/Katasec.DStream.SDK.Core/Katasec.DStream.SDK.Core.csproj

# Build in release mode
/usr/local/share/dotnet/dotnet build dstream-dotnet-sdk.sln -c Release
```

### Running Tests
```bash
# Navigate to the SDK directory
cd ~/progs/dstream/dstream-dotnet-sdk

# Run all tests
/usr/local/share/dotnet/dotnet test dstream-dotnet-sdk.sln

# Run tests for a specific project
/usr/local/share/dotnet/dotnet test tests/Providers.AsbQueue.Tests/Providers.AsbQueue.Tests.csproj

# Run tests with verbose output
/usr/local/share/dotnet/dotnet test dstream-dotnet-sdk.sln -v normal
```

### Modern Provider Development

**Building Sample Providers** (SDK Testing & Examples):
```bash
# Navigate to SDK directory
cd ~/progs/dstream/dstream-dotnet-sdk

# Build counter input provider sample
/usr/local/share/dotnet/dotnet build samples/counter-input-provider/counter-input-provider.csproj -c Release

# Build console output provider sample  
/usr/local/share/dotnet/dotnet build samples/console-output-provider/console-output-provider.csproj -c Release

# Build all samples together
/usr/local/share/dotnet/dotnet build dstream-dotnet-sdk.sln -c Release
```

**✅ VERIFIED Provider Locations (External Pattern):**
- `~/progs/dstream/dstream-counter-input-provider/` - ✅ External repo using NuGet v0.1.1
- `~/progs/dstream/dstream-console-output-provider/` - ✅ External repo using NuGet v0.1.1
- `~/progs/dstream/dstream-dotnet-sdk/samples/` - Legacy examples (for SDK development only)

### Provider Makefile System

Each provider includes a self-documenting Makefile with the following features:

**Self-Documenting Help:**
```bash
# Just run 'make' to see available commands
$ make
build                          Build single self-contained binary
clean                          Remove all build artifacts  
help                           Show available make targets with descriptions
rebuild                        Clean and build from scratch
test                           Test provider with sample config
verify                         Check binary exists in correct location
```

**Available Targets:**
- `make` or `make help` - Show help menu with color-coded descriptions
- `make build` - Build single self-contained binary (~68MB)
- `make clean` - Remove all build artifacts (bin/, obj/, out/)
- `make rebuild` - Clean and build from scratch
- `make verify` - Check binary exists in location expected by dstream.hcl
- `make test` - Test provider with sample configuration

**Key Build Features:**
- **Single-file deployment**: Uses `PublishSingleFile=true` and cleans up extra files
- **Self-contained**: Includes all .NET runtime dependencies
- **Container-ready**: Perfect 68MB binaries for OCI distribution
- **Location-aware**: Places binaries exactly where dstream.hcl expects them

### Legacy Sample Plugin Development
```bash
# Navigate to the legacy sample project (for reference only)
cd ~/progs/dstream/dstream-dotnet-sdk/samples/dstream-dotnet-test

# Build and publish the sample plugin
./build.ps1 publish

# Clean build outputs
./build.ps1 clean
```

### Running Modern Provider Tasks via DStream CLI
```bash
# Navigate to the Go CLI project
cd ~/progs/dstream/dstream

# Run the counter-to-console task (modern provider orchestration)
go run . run counter-to-console

# Run with debug logging
go run . run counter-to-console --log-level debug

# Run with timeout to avoid infinite loops during development
timeout 30s go run . run counter-to-console

# Current dstream.hcl task configuration:
# task "counter-to-console" {
#   type = "providers"  # Modern provider orchestration
#   input {
#     provider_path = "../dstream-dotnet-sdk/samples/counter-input-provider/bin/Release/net9.0/osx-x64/counter-input-provider"
#     config { interval = 1000; max_count = 50 }
#   }
#   output {
#     provider_path = "../dstream-dotnet-sdk/samples/console-output-provider/bin/Release/net9.0/osx-x64/console-output-provider"
#     config { outputFormat = "structured" }
#   }
# }
```

### Package Management
```bash
# Navigate to the SDK directory
cd ~/progs/dstream/dstream-dotnet-sdk

# Restore NuGet packages for all projects
/usr/local/share/dotnet/dotnet restore dstream-dotnet-sdk.sln

# Clean all build outputs
/usr/local/share/dotnet/dotnet clean dstream-dotnet-sdk.sln
```

## Architecture Overview

### Core Components

**✅ SDK Architecture (PUBLISHED & VERIFIED)**
- `Katasec.DStream.Abstractions` v0.1.1: ✅ Core interfaces (`IInputProvider`, `IOutputProvider`, `IInfrastructureProvider`)
- `Katasec.DStream.SDK.Core` v0.1.1: ✅ Base classes (`ProviderBase<TConfig>`, `InfrastructureProviderBase<TConfig>`) 
- `StdioProviderHost`: ✅ Command routing (`RunProviderWithCommandAsync`) for infrastructure lifecycle

**Legacy Architecture (Removed)**
- Legacy components have been removed after successful migration to new SDK

### Plugin Development Pattern

Plugins in this SDK follow a specific pattern:

1. **Config Class**: Defines plugin configuration
```csharp
public sealed record PluginConfig
{
    public int Interval { get; init; } = 5000;
}
```

2. **Provider Implementation**: Inherits from `ProviderBase<TConfig>` and implements provider interfaces
```csharp
public sealed class MyPlugin : ProviderBase<PluginConfig>, IInputProvider
{
    public async IAsyncEnumerable<Envelope> ReadAsync(IPluginContext ctx, CancellationToken ct) { }
}
```

3. **Host Entry Point**: Uses `PluginHost.Run<>()` to bootstrap the plugin
```csharp
await PluginHost.Run<MyPlugin, PluginConfig>();
```

### Provider Types

**Input Providers** (`IInputProvider`)
- Read data from external sources
- Implement `ReadAsync()` returning `IAsyncEnumerable<Envelope>`
- Examples: Counter generators, database CDC, message queues

**Output Providers** (`IOutputProvider`)
- Write data to external destinations  
- Implement `WriteAsync()` accepting `IEnumerable<Envelope>`
- Examples: Console output, Azure Service Bus, databases

### Key Data Types

- `Envelope`: Core data structure with `Payload` (object) and `Meta` (metadata dictionary)
- `IPluginContext`: Runtime context providing logger and services
- `ProviderBase<TConfig>`: Base class handling configuration and context injection

### ✅ **VERIFIED Project Structure (External Provider Pattern)**

```
~/progs/dstream/                              ← ✅ Consolidated project root
├── WARP.md                                  ← ✅ Master context file  
├── dstream/                                 ← ✅ Go CLI orchestrator
│   ├── main.go                             ← ✅ CLI with infrastructure commands
│   ├── dstream.hcl                         ← ✅ Task configuration
│   └── cmd/{init,destroy,plan,status,run}.go ← ✅ All lifecycle commands
├── dstream-dotnet-sdk/                      ← ✅ .NET SDK (publishes to NuGet)
│   ├── sdk/
│   │   ├── Katasec.DStream.Abstractions/   ← ✅ Published v0.1.1
│   │   └── Katasec.DStream.SDK.Core/       ← ✅ Published v0.1.1
│   ├── .github/workflows/publish-nuget.yml ← ✅ Automated publishing
│   ├── VERSION.txt                         ← ✅ v0.1.1
│   └── samples/                            ← SDK testing only
├── dstream-counter-input-provider/          ← ✅ External repo using NuGet v0.1.1
│   ├── Makefile                            ← ✅ Self-documenting build system
│   ├── Program.cs                          ← ✅ StdioProviderHost.RunInputProviderAsync
│   ├── counter-input-provider.csproj       ← ✅ <PackageReference .../>
│   └── out/counter-input-provider           ← ✅ ~68MB single binary
└── dstream-console-output-provider/         ← ✅ External repo using NuGet v0.1.1  
    ├── Writer.cs + Infrastructure.cs       ← ✅ Clean separation of concerns
    ├── console-output-provider.csproj      ← ✅ <PackageReference .../>
    └── out/console-output-provider          ← ✅ ~68MB single binary
```

### Developer Experience

**✅ External Provider Development Pattern:**
```xml
<!-- External providers reference published NuGet packages -->
<PackageReference Include="Katasec.DStream.SDK.Core" Version="0.1.1" />
```

**✅ VERIFIED Working Examples:**
- `dstream-counter-input-provider`: Uses published NuGet v0.1.1 ✅
- `dstream-console-output-provider`: Uses published NuGet v0.1.1 ✅

This enables independent provider development without requiring SDK source code.

## Architectural Decisions

### Provider I/O Model: Stdin/Stdout Streaming over gRPC
er/Writer Model

**Key Insight: One Task = One Command = Three Processes**

DStream uses a **task-centric execution model** where each streaming job runs as an independent task with clean isolation:

```bash
# Each command runs one isolated streaming task
dstream run dotnet-counter    # Demo: Counter → Console
dstream run mssql-cdc         # Production: SQL Server → Azure Service Bus  
dstream run api-monitor       # API polling → Multiple outputs
```

**Task Execution Model:**
```
┌─────────────────────────────────────────────────────────────┐
│  Task: "dotnet-counter"                                     │
│  ┌─────────────┐   ┌─────────────┐   ┌─────────────────┐   │
│  │ DStream CLI │   │ Counter     │   │ Console Output  │   │
│  │ (Process 1) │◄─►│ Input       │◄─►│ Provider        │   │
│  │             │   │ (Process 2) │   │ (Process 3)     │   │
│  └─────────────┘   └─────────────┘   └─────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

**Data Flow: Simple stdin/stdout Piping**
```
┌──────────────┐  stdout   ┌─────────────────┐  stdin   ┌─────────────────┐
│ Input        │  JSON     │ DStream CLI     │  JSON    │ Output          │
│ Provider     │──────────→│ (Orchestrator)  │─────────→│ Provider        │
│              │ envelopes │ • Launch tasks  │envelopes │                 │
│              │           │ • Pipe data     │          │                 │
│              │           │ • Manage procs  │          │                 │
└──────────────┘           └─────────────────┘          └─────────────────┘
```

**Benefits of Task-Based Isolation:**
- **Process Isolation**: Each task runs independently with no shared state
- **Simple I/O Model**: Providers only need to handle stdin/stdout (like a single process with input/output)
- **Easy Testing**: Each provider can be tested independently with shell commands
- **Clean Shutdown**: Killing CLI process cleanly terminates all 3 processes
- **Resource Boundaries**: Clear CPU/memory limits per streaming task
- **Operational Simplicity**: One command = one streaming job (easy monitoring)

### Configuration: HCL-Based Task Definitions

**Decision:** DStream uses a **Reader/Writer abstraction** where input providers are essentially streaming readers and output providers are streaming writers, communicating over gRPC via HashiCorp go-plugin protocol.

**The Key Insight: Native Streaming APIs over gRPC**

This is fundamentally **"Native streaming patterns for structured data over gRPC"** - each language uses its idiomatic streaming abstractions:

**Language-Specific API Mappings:**

**Go (CLI Orchestration):**
```go
// Input Provider = io.Reader pattern
type InputProvider interface {
    Read(ctx context.Context) (<-chan Envelope, error)
}

// Output Provider = io.Writer pattern  
type OutputProvider interface {
    Write(ctx context.Context, envelopes <-chan Envelope) error
}

// Go CLI = Data Pump (like Unix pipe)
func PumpData(reader InputProvider, writer OutputProvider) {
    envelopes, _ := reader.Read(ctx)
    writer.Write(ctx, envelopes)
}
```

**.NET (Provider Implementation):**
```csharp
// Input Provider = IAsyncEnumerable<T> (streaming read)
public interface IInputProvider : IProvider
{
    IAsyncEnumerable<Envelope> ReadAsync(IPluginContext ctx, CancellationToken ct);
    // ↑ Like Stream.ReadAsync() but for structured data
}

// Output Provider = async batch writer
public interface IOutputProvider : IProvider  
{
    Task WriteAsync(IEnumerable<Envelope> batch, IPluginContext ctx, CancellationToken ct);
    // ↑ Like Stream.WriteAsync() but for structured data
}
```

**Runtime Architecture:**
```
┌─────────────────┐    gRPC     ┌─────────────────┐    gRPC     ┌─────────────────┐
│  Input Provider │────────────▶│     Go CLI      │────────────▶│ Output Provider │
│   (.NET Stream) │             │ (io.Reader/     │             │   (.NET Stream) │  
│                 │             │  io.Writer)     │             │                 │
└─────────────────┘             └─────────────────┘             └─────────────────┘
```

**Cross-Language Composability:**
```bash
# Any language can implement providers using native patterns
dotnet-mssql-provider → go-cli → rust-kafka-provider
python-api-provider → go-cli → java-elasticsearch-provider
go-counter-provider → go-cli → dotnet-console-provider
```

**Why this works:**
- **Native patterns:** Each language uses its idiomatic streaming APIs
- **Familiar abstractions:** Go = `io.Reader`/`io.Writer`, .NET = `IAsyncEnumerable`/`WriteAsync`
- **Composable:** Any reader can connect to any writer, regardless of language
- **gRPC abstraction:** Network transport is hidden behind native APIs
- **Battle-tested:** Built on proven streaming patterns from each ecosystem

### Provider Distribution: Independent Binaries

**Decision:** Each input/output provider is an **independent executable binary** distributed as OCI images.

**Options Considered:**

**❌ Option A: Library Loading (NuGet packages)**
- Providers as NuGet packages loaded into single plugin binary
- Complex dependency management
- Coordination required between provider authors
- Difficult ecosystem growth

**✅ Option B: Independent Binaries (Chosen)**
- Each provider is its own executable
- Distributed via OCI container registries
- Zero coordination between provider authors
- Natural ecosystem growth

**Task Configuration (HCL):**
```hcl
# dotnet-counter.hcl - Simple demo task
task "dotnet-counter" {
  input {
    provider_path = "./counter-input-provider"     # Development: local binary
    # provider_href = "ghcr.io/org/counter:v1.0.0"  # Production: OCI image
    config = {
      interval = 1000
      max_count = 100
    }
  }
  
  output {
    provider_path = "./console-output-provider"     # Development: local binary
    # provider_href = "ghcr.io/org/console:v1.0.0"   # Production: OCI image
    config = {
      outputFormat = "structured"
    }
  }
}

# mssql-cdc.hcl - Production CDC task
task "mssql-cdc" {
  input {
    provider_href = "ghcr.io/katasec/mssql-cdc:v1.0.0"  # Production OCI
    config = {
      connection_string = "Server=...;Database=..."
      tables = ["orders", "customers"]
    }
  }
  output {
    provider_href = "ghcr.io/katasec/azure-servicebus:v2.1.0"
    config = {
      connection_string = "Endpoint=sb://..."
      topic_name = "cdc-events"
    }
  }
}
```

**Provider Discovery Evolution:**
- **Phase 1 (Now)**: `provider_path` for local binaries during development
- **Phase 2 (Future)**: `provider_href` for OCI container images in production
- **Configuration**: HCL (not JSON/YAML) - Terraform-style infrastructure-as-code

**Benefits:**
- **Publishing:** Anyone can publish a provider instantly
- **Versioning:** Granular versioning per provider
- **Security:** Audit each provider independently
- **Ecosystem:** No gatekeepers, natural marketplace emerges

**Trade-offs:**
- Higher runtime overhead (2+ processes vs 1)
- More complex CLI orchestration
- Inter-process communication complexity

### Transform Strategy: Queue Chaining

**Decision:** Transforms happen via **queue chaining** rather than separate transform processes.

**Options Considered:**

**❌ Option A: Separate Transform Process**
```
Go CLI → Input Provider → Transform Provider → Output Provider
```
- Too complex (3-process orchestration)
- Complex IPC between all components

**✅ Option B: Embedded Transforms (Chosen)**
- Simple transforms embedded in input/output providers
- Complex transforms via queue chaining

**✅ Option C: Queue Chaining (Chosen)**
```
Input → ASB Queue → Transform Process → ASB Queue → Output
```

**Implementation Patterns:**

**Simple Inline Transforms:**
```csharp
public async IAsyncEnumerable<Envelope> ReadAsync(IPluginContext ctx, CancellationToken ct)
{
    await foreach (var rawEvent in ReadFromSource(ct))
    {
        // Transform before emitting
        var transformed = CleanAndNormalize(rawEvent);
        yield return new Envelope(transformed, metadata);
    }
}
```

**Queue Chaining:**
```hcl
# Stage 1: Raw ingestion
task "ingest" {
  input  { provider_ref = "mssql-cdc" }
  output { provider_ref = "azure-servicebus"; queue = "raw-events" }
}

# Stage 2: Transform
task "transform" {
  input  { provider_ref = "azure-servicebus"; queue = "raw-events" }
  output { provider_ref = "azure-servicebus"; queue = "enriched-events" }
}

# Stage 3: Final destination  
task "sink" {
  input  { provider_ref = "azure-servicebus"; queue = "enriched-events" }
  output { provider_ref = "snowflake-sink" }
}
```

**Benefits:**
- **Fault tolerance:** Queue durability between stages
- **Scalability:** Independent scaling per stage
- **Operations:** Clear monitoring boundaries
- **Testing:** Easy replay and debugging

### Provider Interface Design

**Input Provider Interface:**
```csharp
public interface IInputProvider : IProvider
{
    IAsyncEnumerable<Envelope> ReadAsync(IPluginContext ctx, CancellationToken ct);
}
```

**Output Provider Interface:**
```csharp
public interface IOutputProvider : IProvider
{
    Task WriteAsync(IEnumerable<Envelope> batch, IPluginContext ctx, CancellationToken ct);
}
```

**Core Data Structure:**
```csharp
public readonly record struct Envelope(object Payload, IReadOnlyDictionary<string, object?> Meta);
```

### Modern Provider Development Workflow

**Provider Development (Current - stdin/stdout model):**
1. Create provider using `ProviderBase<TConfig>` and implement `IInputProvider` or `IOutputProvider`
2. Use `StdioProviderHost.RunInputProviderAsync` or `StdioProviderHost.RunOutputProviderAsync`
3. Build as self-contained single-file executable using `make build`
4. Test provider individually using `make test`
5. Package as OCI image for distribution
6. Users reference via `provider_path` (local) or `provider_ref` (OCI) in HCL

**Runtime Workflow (Modern):**
1. `go run . run task-name` - CLI reads dstream.hcl task configuration
2. CLI launches input/output providers as separate processes
3. CLI orchestrates data flow: input.stdout → CLI → output.stdin
4. Providers communicate via JSON envelopes over stdin/stdout pipes
5. Clean shutdown when input provider completes or CLI receives termination signal

**Build System Features:**
- **Self-documenting Makefiles**: `make` shows help menu with available commands
- **Single-file binaries**: ~68MB self-contained executables (perfect for containers)
- **Clean builds**: `make clean` removes all artifacts, `make build` creates only what's needed
- **Verification**: `make verify` checks binary exists in location expected by dstream.hcl
- **Testing**: `make test` runs provider with sample configuration

**Target Ecosystem:**
- **Input Providers:** SQL Server CDC, PostgreSQL CDC, Kafka, REST APIs, File watchers
- **Output Providers:** Azure Service Bus, Amazon SQS, Snowflake, Elasticsearch, webhooks
- **Transform Providers:** Data enrichment, validation, aggregation, ML inference

This architecture enables a "Terraform for data streaming" ecosystem where providers are composable, independently versioned, and community-contributed.

## Architecture Decision: Unix stdin/stdout vs HashiCorp go-plugin (gRPC)

### Decision Context

DStream was initially designed using HashiCorp's go-plugin framework with gRPC communication, chosen for three key factors:
1. **Battle-tested**: Proven in Terraform, Vault, Consul, Packer
2. **Language support**: gRPC works across multiple languages
3. **Performance**: Binary protocol with efficient serialization

However, during development of the independent provider orchestration model, we evaluated **Unix stdin/stdout pipes** as an alternative IPC mechanism.

### Comparison Analysis

| **Factor** | **HashiCorp go-plugin (gRPC)** | **Unix stdin/stdout** | **Winner** |
|------------|--------------------------------|------------------------|------------|
| **🛡️ Battle Tested** | ✅ **Excellent** | ✅ **Legendary** | **stdin/stdout** |
| | • Used by Terraform, Vault, Consul, Packer | • Unix foundation since 1970s | |
| | • Handles process lifecycle, crashes, recovery | • Every shell, container, CI/CD system | |
| | • Plugin discovery and versioning | • Docker, Kubernetes, systemd native | |
| | • 5+ years production at HashiCorp scale | • **50+ years** in production everywhere | |
| **🌐 Language Support** | ✅ **Very Good** | ✅ **Universal** | **stdin/stdout** |
| | • Go (native), .NET, Python, Java, Rust | • **Every programming language** | |
| | • Requires gRPC libraries and proto files | • Built into language standard libraries | |
| | • Need to implement plugin interface | • Just read/write text (JSON/CSV/etc.) | |
| | • Proto compatibility across versions | • No dependencies, just IO streams | |
| **⚡ Performance** | ✅ **Excellent** | 🤔 **Good** | **Depends on use case** |
| | • Binary protocol, efficient serialization | • Text-based JSON parsing overhead | |
| | • Streaming, bidirectional communication | • Sequential pipe processing | |
| | • Type-safe, schema validation | • String parsing/validation needed | |
| | • HTTP/2 multiplexing, compression | • Simple pipe buffering | |

### Critical Performance Factor: Interprocess Communication Latency

| **Metric** | **gRPC (TCP/HTTP2)** | **Unix Pipes (stdin/stdout)** | **Winner** |
|------------|----------------------|-------------------------------|------------|
| **Base Latency** | ~50-200μs | ~1-10μs | **Pipes** |
| **Connection Setup** | TCP handshake + HTTP2 setup | Instant (kernel pipe) | **Pipes** |
| **Memory Copies** | Multiple (network stack) | Single (kernel buffer) | **Pipes** |
| **Context Switches** | User → Kernel → Network → Kernel → User | User → Kernel → User | **Pipes** |
| **Protocol Overhead** | HTTP2 headers + protobuf framing | Raw bytes | **Pipes** |
| **Throughput** | ~100K-500K msgs/sec | ~1M+ msgs/sec | **Pipes** |

**Key Insight**: Unix pipes provide **10-50x faster interprocess communication** than gRPC for local process orchestration.

### Real-World DStream Scenarios

#### Scenario 1: Azure Activity Logs → Azure Service Bus
```hcl
task "activity-logs-to-servicebus" {
  type = "providers"
  
  input {
    provider_path = "./azure-activity-logs-provider"
    config = {
      subscription_id = "12345678-1234-1234-1234-123456789012"
      resource_groups = ["production", "staging"]  
      polling_interval = "30s"
    }
  }
  
  output {
    provider_path = "./azure-servicebus-provider"
    config = {
      connection_string = "Endpoint=sb://..."
      queue_name = "activity-logs"
      batch_size = 100
    }
  }
}
```

**Data Flow**:
```
Azure Activity Logs API → Input Provider → stdout → CLI → stdin → Output Provider → Service Bus Queue
```

**Testing Capability**:
```bash
# Test input provider independently:
echo '{"subscription_id":"...","polling_interval":"30s"}' | ./azure-activity-logs-provider

# Test output provider independently:  
echo '{"connection_string":"...","queue_name":"test"}' | ./azure-servicebus-provider

# Test full pipeline manually:
echo '{"subscription_id":"..."}' | ./azure-activity-logs-provider | ./azure-servicebus-provider
```

#### Scenario 2: MS SQL CDC → Azure Data Factory
```hcl
task "mssql-cdc-to-adf" {
  type = "providers"
  
  input {
    provider_path = "./mssql-cdc-provider"
    config = {
      connection_string = "Server=sql.company.com;Database=Orders;..."
      tables = ["Orders", "Customers", "OrderItems"]
      cdc_start_lsn = "auto"
      polling_interval = "5s"
      batch_size = 1000
    }
  }
  
  output {
    provider_path = "./azure-datafactory-provider"  
    config = {
      subscription_id = "12345678-1234-1234-1234-123456789012"
      resource_group = "data-engineering"
      factory_name = "company-adf"
      pipeline_name = "ingest-cdc-changes"
    }
  }
}
```

**High-Frequency CDC Example**:
```bash
# SQL Server CDC processing 10,000 changes/second:

# With gRPC:
10,000 msgs × 100μs latency = 1 second just in IPC overhead
+ Processing time = significant bottleneck

# With Pipes:  
10,000 msgs × 5μs latency = 50ms in IPC overhead
+ Processing time = IPC negligible
```

### Final Architecture Decision: Unix stdin/stdout

**Decision**: DStream adopts **Unix stdin/stdout pipes** for independent provider orchestration.

**Rationale**:
1. **Superior IPC Performance**: 10-50x faster interprocess communication
2. **Universal Language Support**: Every programming language supports stdin/stdout natively
3. **Legendary Battle Testing**: 50+ years of production use in Unix systems
4. **Perfect Alignment**: Matches DStream's vision as "Unix pipeline for data"
5. **Developer Experience**: Trivial testing with shell commands
6. **Operational Simplicity**: Standard Unix tooling (pipes, redirects, etc.)

**Trade-offs Accepted**:
- Text-based JSON parsing overhead vs binary protobuf (negligible for I/O-bound workloads)
- Manual schema validation vs automatic protobuf validation
- Simple sequential processing vs complex bidirectional communication

**Implementation Model**:
```bash
# Each provider is a standalone binary:
echo 'CONFIG_JSON' | ./input-provider | ./output-provider

# CLI orchestrates the pipeline:
dstream run task-name
# → CLI launches input-provider and output-provider
# → CLI pipes: input.stdout → output.stdin
# → CLI handles process lifecycle and graceful shutdown
```

**This architectural decision enables DStream to fulfill its vision as a "Unix pipeline for data" with maximum simplicity, performance, and universal language support.**

## Current Architecture Status & Evolution Plan

### Background

DStream started with SQL Server CDC embedded in the Go CLI, then evolved to support Go plugins (like [dstream-ingester-mssql](https://github.com/katasec/dstream-ingester-mssql)). The .NET plugin support was added to enable .NET developer teams to contribute to the ecosystem.

### Current State (Working)

**✅ Go CLI ↔ .NET Plugin Communication**
- gRPC communication via HashiCorp go-plugin protocol works
- Configuration passing from HCL → Go CLI → .NET plugin works
- Basic .NET counter plugin runs successfully

**❌ .NET Output Provider Routing (Broken)**
- Current `PluginServiceImpl.cs` only handles input providers
- Output configuration is received but ignored
- `ctx.Emit()` just logs instead of routing to output providers

### Architecture Evolution Plan

**Phase 1: Fix .NET Plugin Architecture (Current)**
```
Go CLI → .NET Plugin Process
             ↓
         (Input + Output routing)
```

**Immediate Goals:**
1. Fix .NET `PluginServiceImpl` to parse output provider config
2. Implement provider registry/factory pattern
3. Route data: Input Provider → Output Provider (within same process)
4. Get console output working with counter input

**Phase 2: Separate .NET Provider Binaries (Future)**
```
Go CLI → Input Provider Binary (.NET)
       ↓
       → Output Provider Binary (.NET)
```

**Future Goals:**
1. Evolve Go CLI to orchestrate separate input/output processes
2. Create provider templates and OCI distribution
3. Build ecosystem of independent provider binaries

### Current Implementation Priority

**Step 1: Fix Output Provider Routing**
- Parse `StartRequest.Output` configuration
- Instantiate appropriate output provider (ConsoleOutputProvider)
- Route `ctx.Emit()` to output provider instead of logging

**Step 2: Validate Input Provider Pattern**
- Ensure input providers work correctly in new architecture
- Test with counter and future SQL Server CDC provider

**Step 3: Build Provider Ecosystem**
- Create MSSQL CDC input provider
- Create Azure Service Bus output provider
- Document provider development patterns

### Development Practices

**Critical: Incremental Changes with Validation**

Every change must be validated to ensure we don't break the working communication:

1. **Make localized changes** - small, focused modifications
2. **Compile and test** after each change:
   ```bash
   # Build the plugin
   cd ~/progs/dstream/dstream-dotnet-sdk/samples/dstream-dotnet-test
   pwsh -c "./build.ps1 clean && ./build.ps1 publish"
   
   # Test end-to-end communication
   cd ~/progs/dstream/dstream
   go run . run counter-to-console
   
   # Verify: Should see counter data flowing from providers → Go CLI
   ```

### Complete Development Workflow (Modern Providers)

**Step 1: Build Providers**
```bash
# Build input provider
cd ~/progs/dstream/dstream-counter-input-provider
make clean && make build

# Build output provider
cd ~/progs/dstream/dstream-console-output-provider
make clean && make build

# Verify both providers are built
make verify
```

**Step 2: Test Individual Providers**
```bash
# Test input provider (generates 3 messages)
cd ~/progs/dstream/dstream-counter-input-provider
echo '{"interval": 1000, "max_count": 3}' | bin/Release/net9.0/osx-x64/counter-input-provider

# Test output provider (processes sample data)
cd ~/progs/dstream/dstream-console-output-provider
echo '{"outputFormat": "simple"}' | bin/Release/net9.0/osx-x64/console-output-provider
```

**Step 3: Test Complete Pipeline**
```bash
cd ~/progs/dstream/dstream
go run . run counter-to-console

# Or with timeout for development
timeout 30s go run . run counter-to-console
```

**Step 4: Development Cycle**
```bash
# Make changes to provider code, then:
cd ~/progs/dstream/dstream-counter-input-provider
make rebuild           # Clean build
make verify           # Check binary location

# Test the updated pipeline
cd ~/progs/dstream/dstream
go run . run counter-to-console
```
3. **Validate data flow** - ensure counter data still flows correctly
4. **Only proceed** if the basic communication still works

**✅ VERIFIED WORKING ARCHITECTURE:**
- ✅ Go CLI with infrastructure lifecycle commands (init/destroy/plan/status/run)
- ✅ External providers using published NuGet packages (v0.1.1)
- ✅ stdin/stdout communication with command routing via JSON envelopes
- ✅ OCI distribution working with both provider_path and provider_ref
- ✅ Infrastructure lifecycle management with `IInfrastructureProvider`
- ✅ Complete end-to-end workflow: counter input → console output

**Current Architecture Status**: Foundation Phase 0, 1 & 2 COMPLETE

### Integration with DStream CLI

The DStream CLI is a Go application located at `~/progs/dstream/dstream` that serves as the orchestrator for .NET providers.

### Existing Go-based Providers

**MSSQL CDC Provider**: There's already a working Go-based MSSQL CDC provider at `~/progs/dstream-ingester-mssql` ([GitHub](https://github.com/katasec/dstream-ingester-mssql)) that:
- Uses HashiCorp go-plugin protocol over gRPC
- Handles SQL Server Change Data Capture
- Works with the current DStream CLI
- Will continue to work alongside .NET providers (language-agnostic ecosystem)

This means the initial .NET provider focus should be on:
- **Testing providers** (counter input, console output)
- **Complementary providers** (Azure Service Bus output)
- **Validating cross-language compatibility** (Go input → .NET output)

**Plugin Lifecycle:**
1. Go CLI parses `dstream.hcl` configuration file
2. CLI launches .NET plugin executable as subprocess using `exec.Command()`
3. .NET plugin starts gRPC server and outputs handshake: `1|1|tcp|127.0.0.1:{port}|grpc`
4. Go CLI connects to plugin's gRPC server
5. CLI sends `StartRequest` with config, input, and output provider settings
6. Plugin runs until cancelled by CLI

**gRPC Interface (defined in `proto/plugin.proto`):**
```protobuf
service Plugin {
  rpc GetSchema (google.protobuf.Empty) returns (GetSchemaResponse);
  rpc Start (StartRequest) returns (google.protobuf.Empty);
}
```

**Configuration Flow:**
- HCL config → Go CLI → JSON → gRPC `StartRequest` → .NET deserialization → Plugin config
- Configuration includes global plugin settings, input provider config, and output provider config

### Development Notes

**Technical Requirements:**
- Plugins must target .NET 9.0 or later
- Use `PublishSingleFile=true` for deployment to create standalone executables
- All plugins must implement gRPC server using ASP.NET Core + Kestrel (HTTP/2)
- Plugins communicate exclusively via gRPC (HashiCorp go-plugin protocol)

**Logging Integration:**
- HCLogger (from HCLog.Net) is used for logging integration with HashiCorp tools
- Logs are written to stderr (stdout is reserved for handshake protocol)
- Log format is compatible with go-hclog JSON structure

**Configuration System:**
- Configuration is automatically bound from HCL → JSON → .NET config objects
- Uses `google.protobuf.Struct` for config transport over gRPC
- Plugin receives global config, input config, and output config separately
- The `[EnumeratorCancellation]` attribute is required on cancellation tokens in async enumerables

**Plugin Detection:**
- CLI detects plugins via environment variables: `PLUGIN_PROTOCOL_VERSIONS`, `PLUGIN_MIN_PORT`
- Direct execution shows HashiCorp-style warning message

## Common Provider Patterns

### Input Provider Template
```csharp
public sealed class MyInputProvider : ProviderBase<MyConfig>, IInputProvider
{
    public async IAsyncEnumerable<Envelope> ReadAsync(IPluginContext ctx, 
        [EnumeratorCancellation] CancellationToken ct)
    {
        var log = (HCLogger)ctx.Logger;
        
        while (!ct.IsCancellationRequested)
        {
            // Read data from source
            var data = await ReadFromSource(ct);
            
            var meta = new Dictionary<string, object?> { ["source"] = "mysource" };
            yield return new Envelope(data, meta);
        }
    }
}
```

### Output Provider Template
```csharp
public sealed class MyOutputProvider : ProviderBase<MyConfig>, IOutputProvider
{
    public Task WriteAsync(IEnumerable<Envelope> batch, IPluginContext ctx, CancellationToken ct)
    {
        var log = (HCLogger)ctx.Logger;
        
        foreach (var envelope in batch)
        {
            if (ct.IsCancellationRequested) break;
            // Write envelope.Payload to destination
            WriteToDestination(envelope.Payload, envelope.Meta, ct);
        }
        
        return Task.CompletedTask;
    }
}
```
