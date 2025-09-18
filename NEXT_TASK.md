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

### Phase 3: Database Table-Aware Azure Service Bus Provider

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

- [ ] CLI accepts `init`, `destroy`, `plan`, `status` commands
- [ ] Commands are routed to providers via JSON command header
- [ ] ASB output provider creates/destroys queues dynamically
- [ ] End-to-end test: MS SQL CDC tables → ASB queues with full lifecycle
- [ ] Backward compatibility maintained for existing providers

## 📖 Reference

This implements the "Terraform for data streaming" vision with infrastructure-as-code embedded directly in providers while maintaining the elegant Unix stdin/stdout pipeline architecture.

---

**When ready to continue, start with Phase 1: CLI Infrastructure Commands** ⭐