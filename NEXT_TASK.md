# 🎯 NEXT TASK: Infrastructure Lifecycle Management for Output Providers

## 📋 Task Summary

Implement infrastructure lifecycle management to support the **MS SQL CDC → Azure Service Bus** scenario where:

- **Input**: MS SQL CDC monitors multiple tables (e.g., `Persons`, `Orders`, `Customers`)
- **Output**: Azure Service Bus creates dedicated queues per table (e.g., `Persons_cdc_events`, `Orders_cdc_events`)
- **Need**: Create/destroy ASB queues dynamically based on monitored tables

## 🎯 Core Challenge Solved

**Verb Routing**: Pass lifecycle commands (`init`/`run`/`destroy`/`plan`/`status`) to providers through existing stdin/stdout protocol.

**Solution**: Extend JSON config with command header:

```json
{
  "command": "init",  // New field for lifecycle verbs
  "config": {
    "connection_string": "...",
    "tables": ["Persons", "Orders", "Customers"]
  }
}
```

## 📚 Design Documents

- [`DESIGN_NOTES.md`](./DESIGN_NOTES.md) - Complete infrastructure lifecycle design
- [`DESIGN_NOTES_VERB_ROUTING.md`](./DESIGN_NOTES_VERB_ROUTING.md) - Detailed verb routing implementation

## 🏗️ Implementation Plan

### Phase 1: CLI Infrastructure Commands ⭐ **START HERE**

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

### Phase 1: CLI Infrastructure Commands
- [ ] CLI accepts `init`, `destroy`, `plan`, `status` commands
- [ ] Commands are routed to providers via JSON command header
- [ ] Backward compatibility maintained for existing providers

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

### Complete Pipeline Success
- [ ] **Full workflow**: `dstream init mssql-to-asb` → `dstream run mssql-to-asb` → `dstream destroy mssql-to-asb`
- [ ] **Data flow**: SQL CDC table changes → JSON envelopes → Table-specific ASB queues
- [ ] **Infrastructure**: Dynamic queue creation/destruction based on monitored tables

## 📖 Reference

This implements the "Terraform for data streaming" vision with infrastructure-as-code embedded directly in providers while maintaining the elegant Unix stdin/stdout pipeline architecture.

---

**When ready to continue, start with Phase 1: CLI Infrastructure Commands** ⭐