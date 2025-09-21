# 🎯 DStream Development Roadmap

## 📋 Task Summary

**Current Status**: Infrastructure lifecycle management (Phases 1-2) is ✅ **COMPLETE** per `DESIGN_NOTES_PHASE_2_COMPLETE.md`.

**Next Priority**: Establish **SDK publishing pipeline and clean repository structure** to enable:

1. **External provider development** - Independent repos consuming published SDK NuGet packages
2. **Clean architecture** - Archive legacy `providers/` folder, rely on production-ready `samples/`
3. **Ecosystem growth** - Enable community provider development with proper tooling

## 🎯 Foundation Challenge

**Repository Structure**: Current mixed approach with both `providers/` (legacy SDK-style) and `samples/` (production executable) folders creates confusion.

**Solution**: Simplify the project structure:

```bash
# Archive legacy structure
git mv providers providers-archived

# Single source of truth for sample providers
samples/
├── counter-input-provider/     # ✅ Production-ready executable
├── console-output-provider/    # ✅ Production-ready executable
└── Playground/                 # ✅ SDK testing sandbox

# Future external development (after NuGet publishing)
External repos will reference SDK packages from NuGet
```

## 📚 Design Documents

- [`DESIGN_NOTES.md`](./DESIGN_NOTES.md) - Complete infrastructure lifecycle design
- [`DESIGN_NOTES_VERB_ROUTING.md`](./DESIGN_NOTES_VERB_ROUTING.md) - Detailed verb routing implementation

## 🏗️ Implementation Plan

### Phase 0: Repository Structure & SDK Publishing ⭐ **FOUNDATION** 

**Goal**: Establish proper SDK publishing pipeline and clean project structure to enable external provider development.

#### 0.1: SDK Repository Cleanup ✅ **COMPLETED**
```bash
# Clean project structure achieved:
~/progs/dstream/
├── dstream/                          # Go CLI orchestrator
├── dstream-dotnet-sdk/              # .NET SDK
│   ├── sdk/                         # Core SDK packages
│   ├── samples/                     # ✅ Primary provider examples
│   │   ├── counter-input-provider/      # Working executable
│   │   ├── console-output-provider/     # Working executable
│   │   └── Playground/                  # SDK testing
│   └── tests/                       # Unit tests
└── dstream-ingester-mssql/           # Existing Go provider
```

**Completed Actions**: 
- ❌ Deleted duplicate standalone provider directories
- ❌ Removed legacy `providers/` folder with placeholder implementations
- ✅ Single source of truth: `samples/` contains working provider examples
- ✅ Clean separation between SDK framework (`sdk/`) and provider examples (`samples/`)

#### 0.2: NuGet Publishing Automation
```bash
# Files to create in dstream-dotnet-sdk/:
# .github/workflows/publish-nuget.yml    # GitHub Actions for automated publishing
# scripts/pack-and-publish.ps1           # PowerShell script for local publishing
# scripts/version-bump.ps1               # Semantic versioning management
# VERSION.txt                            # Central version management
```

**Publishing Pipeline Features**:
- **Semantic versioning**: Automated version bumping (major.minor.patch)
- **Tag-triggered releases**: `git tag v1.2.3` → automatic NuGet publish
- **Pre-release support**: `v1.2.3-beta.1` for development versions
- **Multi-package coordination**: All SDK packages versioned together
- **Release notes**: Auto-generated from git commits and PR descriptions

#### 0.3: External Provider Repository Template
```bash
# Create template repository structure:
~/progs/
├── dstream/                                    # Main orchestrator
├── dstream-dotnet-sdk/                        # SDK source (publishes to NuGet)
└── dstream-providers/                         # NEW - External repo using NuGet packages
    ├── counter-input-provider/                # Copy from samples, uses NuGet
    ├── console-output-provider/               # Copy from samples, uses NuGet
    ├── .github/workflows/publish-oci.yml     # Automated OCI publishing
    └── README.md                              # External provider development guide
```

#### 0.4: OCI Container Distribution Validation ⭐ **END-TO-END PROOF**

**Goal**: Prove the entire ecosystem works by building and running providers from OCI containers.

**Steps**:
1. **Create `dstream-providers` repository**:
   ```bash
   # New external repository
   mkdir ~/progs/dstream-providers
   cd ~/progs/dstream-providers
   git init
   ```

2. **Copy sample providers to use NuGet packages**:
   ```bash
   # Copy samples but reference published NuGet packages instead of local projects
   cp -r ../dstream-dotnet-sdk/samples/counter-input-provider .
   cp -r ../dstream-dotnet-sdk/samples/console-output-provider .
   
   # Update .csproj files to use NuGet package references
   # <PackageReference Include="Katasec.DStream.SDK.Core" Version="0.1.0" />
   ```

3. **Create cross-platform OCI build system**:
   ```dockerfile
   # Multi-stage Dockerfile for cross-platform builds
   FROM mcr.microsoft.com/dotnet/sdk:9.0 AS build
   WORKDIR /src
   COPY . .
   RUN dotnet publish -c Release -o /app --self-contained --runtime linux-x64
   
   FROM mcr.microsoft.com/dotnet/runtime:9.0-alpine AS runtime
   WORKDIR /app
   COPY --from=build /app .
   ENTRYPOINT ["./provider"]
   ```

4. **Create GitHub Actions for OCI publishing**:
   ```yaml
   # .github/workflows/publish-oci.yml
   name: Publish OCI Images
   on:
     push:
       tags: ['v*']
   jobs:
     build-and-push:
       runs-on: ubuntu-latest
       strategy:
         matrix:
           provider: [counter-input-provider, console-output-provider]
       steps:
         - uses: actions/checkout@v4
         - name: Build and push OCI image
           run: |
             docker build -t ghcr.io/katasec/${{ matrix.provider }}:${{ github.ref_name }} \
               --platform linux/amd64,linux/arm64,windows/amd64 \
               ${{ matrix.provider }}/
   ```

5. **Update dstream.hcl to use OCI images**:
   ```hcl
   task "oci-validation" {
     type = "providers"
     
     input {
       provider_image = "ghcr.io/katasec/counter-input-provider:v0.1.0"
       config = { interval = 1000; max_count = 5 }
     }
     
     output {
       provider_image = "ghcr.io/katasec/console-output-provider:v0.1.0"
       config = { outputFormat = "simple" }
     }
   }
   ```

6. **Extend DStream CLI to support OCI images**:
   ```go
   // pkg/executor/oci.go
   func (e *Executor) pullAndRunOCIProvider(imageRef string, config map[string]interface{}) error {
       // docker pull ghcr.io/katasec/counter-input-provider:v0.1.0
       // docker run --rm -i ghcr.io/katasec/counter-input-provider:v0.1.0
   }
   ```

**End-to-End Validation**:
```bash
# Test the complete OCI workflow
cd ~/progs/dstream/dstream
go run . run oci-validation

# Should:
# 1. Pull OCI images automatically
# 2. Run containers with JSON config via stdin
# 3. Pipe data between containerized providers
# 4. Display results successfully
```

**Tasks**:
- [ ] Create automated NuGet publishing GitHub Actions workflow
- [ ] Implement semantic versioning with central VERSION.txt management
- [x] ✅ **COMPLETED** - Remove duplicate provider directories and clean up repository structure
- [x] ✅ **COMPLETED** - Update solution file to remove deleted provider project references  
- [ ] **Create `dstream-providers` repository** with external provider examples
- [ ] **Copy sample providers to use NuGet packages** instead of project references
- [ ] **Create cross-platform OCI build system** with Dockerfiles and GitHub Actions
- [ ] **Extend DStream CLI to support `provider_image`** alongside `provider_path`
- [ ] **End-to-end OCI validation**: Pull and run providers from container registry
- [ ] Document external provider development pattern
- [ ] Test complete external provider development workflow

### Phase 1: CLI Infrastructure Commands ⭐ **CURRENT FOCUS**

```bash
# New CLI commands to implement
dstream init mssql-to-asb      # Provision infrastructure for task
dstream plan mssql-to-asb      # Show what would be created/destroyed
dstream run mssql-to-asb       # Run the data pipeline (existing)
dstream status mssql-to-asb    # Show current infrastructure state
dstream destroy mssql-to-asb   # Clean up infrastructure for task
```

**Files to modify:**
- [ ] `cmd/` - Add new CLI commands (`init.go`, `destroy.go`, `plan.go`, `status.go`)
- [ ] `pkg/executor/executor.go` - Add command routing to `ExecuteProviderTask(task, command)`
- [ ] `pkg/executor/providers.go` - Extend to send command in JSON config

### Phase 2: .NET SDK Extensions

**Files to create/modify:**
- [ ] Add `IInfrastructureProvider` interface to `Katasec.DStream.Abstractions`
- [ ] Extend `StdioProviderHost` to handle command routing
- [ ] Create `CommandEnvelope<TConfig>` for deserialization
- [ ] Add `InfrastructureProviderBase<TConfig>` with Pulumi integration

### Phase 3: SQL Server CDC Input Provider Extraction ⭐ **HIGH VALUE**

**Extract production-tested Go SQL CDC code and convert protocol:**

#### Repository Setup
- [ ] Checkout earlier DStream version with embedded SQL CDC
- [ ] Create new directory: `~/progs/dstream/sqlcdc-input-provider`
- [ ] Initialize as separate Go module: `go mod init github.com/katasec/sqlcdc-input-provider`
- [ ] Extract production SQL CDC business logic from embedded CLI version

#### Business Logic to Preserve ✅ (Keep As-Is)
- [ ] SQL Server connection management and pooling
- [ ] CDC table discovery and monitoring loops
- [ ] LSN (Log Sequence Number) tracking and offset management
- [ ] Change record parsing and transformation
- [ ] Retry logic and error handling strategies
- [ ] Distributed locking (Azure Blob Storage integration)
- [ ] Adaptive polling and backoff strategies
- [ ] Production-tested CDC change detection

#### Integration Protocol to Convert 🔄 (gRPC → stdin/stdout)
- [ ] **Remove**: gRPC server setup and HashiCorp plugin handshake
- [ ] **Remove**: Protobuf message definitions and `StartRequest`/`GetSchema` methods
- [ ] **Remove**: gRPC streaming and plugin lifecycle management
- [ ] **Add**: JSON configuration reading from stdin (first line)
- [ ] **Add**: JSON envelope writing to stdout (continuous stream)
- [ ] **Add**: Error logging to stderr and graceful SIGTERM shutdown

#### Provider Protocol Implementation
- [ ] **Config Protocol**: Read JSON config from stdin: `{"connection_string": "...", "tables": ["Orders", "Customers"]}`
- [ ] **Data Protocol**: Write JSON envelopes to stdout:
  ```json
  {"data": {"ID": "123", "Name": "Ameer"}, "metadata": {"TableName": "Persons", "OperationType": "Insert", "LSN": "0000004c000028200003"}}
  ```
- [ ] **Testing**: Validate with `echo '{"tables":["TestTable"]}' | ./sqlcdc-input-provider`

#### Provider Naming
- **Binary**: `dstream-sqlcdc-input-provider`
- **Directory**: `sqlcdc-input-provider`
- **Module**: `github.com/katasec/sqlcdc-input-provider`

#### Extraction Workflow (Step-by-Step)
```bash
# 1. Find the earlier version with embedded SQL CDC
cd ~/progs/dstream/dstream
git log --oneline --grep="CDC" --grep="sql" --all  # Find relevant commits

# 2. Checkout that version
git checkout <commit-with-embedded-cdc>

# 3. Create new provider directory
cd ~/progs/dstream
mkdir sqlcdc-input-provider
cd sqlcdc-input-provider

# 4. Initialize new Go module
go mod init github.com/katasec/sqlcdc-input-provider

# 5. Copy relevant CDC packages
# Copy from: dstream/pkg/cdc/, dstream/internal/sqlcdc/, etc.
# Inspect and identify which packages contain the CDC business logic

# 6. Create main.go with stdin/stdout protocol
# Replace gRPC interface with JSON stdin/stdout handling

# 7. Test extraction
echo '{"connection_string":"test", "tables":["TestTable"]}' | go run .
```

#### Key Code Transformation Pattern
```go
// OLD: gRPC plugin pattern
func (p *CDCPlugin) Start(ctx context.Context, req *pb.StartRequest) error {
    // Business logic here
    for change := range p.monitorChanges() {
        p.sendViaGRPC(change)  // Remove this
    }
}

// NEW: stdin/stdout provider pattern  
func main() {
    config := readJSONFromStdin()  // Add this
    provider := NewCDCProvider(config)
    
    for change := range provider.monitorChanges() {  // Keep business logic
        envelope := createJSONEnvelope(change)    // Add this
        writeJSONToStdout(envelope)               // Add this
    }
}
```

### Phase 4: Database Table-Aware Azure Service Bus Provider

**New provider to create:**
- [ ] `DbtableAsbOutputProvider` implementing both `IOutputProvider` and `IInfrastructureProvider`
- [ ] Embedded Pulumi stack for ASB queue management  
- [ ] Dynamic queue creation based on database table metadata from envelopes
- [ ] Table-aware queue naming: `{TableName}_cdc_events`
- [ ] Compatible with any tabular input provider (SQL Server CDC, PostgreSQL CDC, MySQL CDC, etc.)

### Phase 5: OCI Container Distribution & Production Ecosystem

**Goal**: Enable production deployment with OCI container distribution and advanced orchestration features.

#### 5.1: Container Build System
```bash
# Add to each provider repository:
# Dockerfile                    # Multi-stage build with .NET 9 runtime
# .github/workflows/build.yml   # Automated OCI image building
# docker-compose.yml            # Local testing environment
```

**Container Architecture**:
```dockerfile
# Example provider Dockerfile
FROM mcr.microsoft.com/dotnet/runtime:9.0-alpine AS runtime
COPY bin/Release/net9.0/linux-x64/mssql-cdc-provider /app/provider
ENTRYPOINT ["/app/provider"]
# Result: ~75MB container with provider + .NET runtime
```

#### 5.2: HCL Provider References Evolution
```hcl
# Phase 1: Local development (current)
task "mssql-to-asb" {
  input {
    provider_path = "../dstream-mssql-cdc-provider/bin/Release/net9.0/osx-x64/mssql-cdc-provider"
  }
}

# Phase 2: OCI container distribution (future)
task "mssql-to-asb" {
  input {
    provider_image = "ghcr.io/katasec/mssql-cdc-provider:v2.1.0"
  }
  output {
    provider_image = "ghcr.io/katasec/asb-output-provider:v1.3.0"
  }
}
```

#### 5.3: Provider Marketplace Ecosystem
**Registry Structure**:
```
ghcr.io/katasec/                       # Official providers
├── mssql-cdc-provider:v2.1.0         # SQL Server CDC input
├── postgres-cdc-provider:v1.0.0      # PostgreSQL CDC input  
├── asb-output-provider:v1.3.0        # Azure Service Bus output
└── snowflake-output-provider:v1.0.0  # Snowflake data warehouse output

ghcr.io/community/                     # Community providers
├── kafka-input-provider:v0.9.0       # Community Kafka input
├── elasticsearch-output:v1.1.0       # Community Elasticsearch output
└── webhook-output-provider:v0.5.0    # Community webhook notifications
```

#### 5.4: Advanced CLI Features
```bash
# Provider discovery and management
dstream providers list                           # Show available providers
dstream providers search mssql                   # Search provider registry
dstream providers pull ghcr.io/katasec/mssql-cdc-provider:v2.1.0

# Task management with containers
dstream init mssql-to-asb --pull-images        # Auto-pull required images
dstream run mssql-to-asb --detach              # Background execution
dstream status mssql-to-asb --detailed         # Detailed status with container info
```

**Tasks**:
- [ ] Implement OCI container build system for all providers
- [ ] Create automated container publishing pipeline
- [ ] Extend CLI to support `provider_image` alongside `provider_path`
- [ ] Implement container orchestration (pull, run, cleanup)
- [ ] Build provider registry and discovery system
- [ ] Create provider marketplace documentation
- [ ] Implement advanced CLI provider management commands

## 💡 Key Design Decisions Made

### ✅ **Embedded Pulumi** (Not External Terraform)
- Pulumi code embedded directly in provider binary
- Ships in same OCI image with provider
- Infrastructure and code versions stay synchronized

### ✅ **Command Header in JSON Config** (Not CLI args or env vars)
- Extends existing stdin/stdout protocol
- Maintains Unix pipeline philosophy
- Works with any programming language

### ✅ **Interface-Based Provider Design**
```csharp
// Database table-aware ASB provider implements both interfaces
public class DbtableAsbOutputProvider : InfrastructureProviderBase<DbtableAsbConfig>, IOutputProvider
{
    // Infrastructure methods: InitializeAsync(), DestroyAsync(), PlanAsync()
    // Data methods: WriteAsync() - routes based on TableName metadata
}
```

### ✅ **Task-Level Lifecycle Management**
- CLI operates on complete tasks (not individual providers)
- Orchestrates both input and output provider infrastructure
- Clean separation between infrastructure and data operations

## 🎪 Example Scenario

```hcl
# dstream.hcl
task "mssql-to-asb" {
  type = "providers"
  
  input {
    provider_path = "./mssql-cdc-provider"
    config {
      connection_string = "{{ env \"SQL_CONNECTION\" }}"
      tables = ["Persons", "Orders", "Customers"]
    }
  }
  
  output {
    provider_path = "./dstream-dbtable-asb-output-provider"
    config {
      connection_string = "{{ env \"ASB_CONNECTION\" }}"
    }
  }
}
```

```bash
# Workflow
dstream init mssql-to-asb     # Creates: Persons_cdc_events, Orders_cdc_events, Customers_cdc_events
dstream run mssql-to-asb      # Streams CDC data to created queues
dstream destroy mssql-to-asb  # Cleans up all created queues
```

## 🎯 Success Criteria

### Phase 0: Repository Structure & SDK Publishing ⭐ **FOUNDATION**
- [ ] `providers/` folder archived to `providers-archived/`
- [ ] `samples/` established as primary provider examples
- [ ] Automated NuGet publishing pipeline with GitHub Actions
- [ ] Semantic versioning with central VERSION.txt management
- [ ] External provider template consuming published SDK NuGet packages
- [ ] End-to-end validation: external provider development workflow

### Phase 1: CLI Infrastructure Commands ✅ **COMPLETED**
- [x] CLI accepts `init`, `destroy`, `plan`, `status` commands
- [x] Commands are routed to providers via JSON command header
- [x] Backward compatibility maintained for existing providers
- [x] **Status**: Phase 1 and Phase 2 are architecturally complete per `DESIGN_NOTES_PHASE_2_COMPLETE.md`

### Phase 3: SQL CDC Provider Extraction
- [ ] Production SQL CDC logic extracted and preserved
- [ ] Provider reads JSON config from stdin, writes JSON envelopes to stdout
- [ ] Compatible with CLI stdin/stdout orchestration protocol
- [ ] Independent testing: `echo '{"tables":["TestTable"]}' | ./sqlcdc-input-provider`
- [ ] All CDC features working: table discovery, LSN tracking, change detection

### Phase 4: Database Table-Aware ASB Provider
- [ ] ASB output provider creates/destroys queues dynamically based on table metadata
- [ ] Infrastructure lifecycle management with embedded Pulumi
- [ ] End-to-end test: SQL CDC tables → ASB queues with full lifecycle

### Phase 5: OCI Container Distribution & Production Ecosystem
- [ ] OCI container build system for all providers
- [ ] CLI supports both `provider_path` (local) and `provider_image` (OCI)
- [ ] Provider marketplace with discovery and management commands
- [ ] Production-ready container orchestration and deployment

### Complete Pipeline Success
- [ ] **Full workflow**: `dstream init mssql-to-asb` → `dstream run mssql-to-asb` → `dstream destroy mssql-to-asb`
- [ ] **Data flow**: SQL CDC table changes → JSON envelopes → Table-specific ASB queues
- [ ] **Infrastructure**: Dynamic queue creation/destruction based on monitored tables
- [ ] **Ecosystem**: Community providers via OCI container distribution

## 📖 Reference

This implements the "Terraform for data streaming" vision with infrastructure-as-code embedded directly in providers while maintaining the elegant Unix stdin/stdout pipeline architecture.

---

**When ready to continue, start with Phase 1: CLI Infrastructure Commands** ⭐