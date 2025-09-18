# DStream Design Notes & TODOs

## Infrastructure Lifecycle Management (Future)

### Problem Statement
- MS SQL CDC with 10 tables → 10 dedicated ASB queues
- Need infrastructure as code (Pulumi) for resource lifecycle
- Providers need to provision/destroy their required infrastructure
- Must work with stdin/stdout architecture (not gRPC)

### Proposed Solution: Provider Lifecycle Commands

#### CLI Commands
```bash
dstream init <taskname>      # Provision infrastructure for task
dstream destroy <taskname>   # Clean up infrastructure for task  
dstream plan <taskname>      # Show what infrastructure would be created/destroyed
dstream status <taskname>    # Show current infrastructure state
dstream run <taskname>       # Run the actual data pipeline (existing)
```

#### Provider Interface Extension
- Extend stdin/stdout protocol to support lifecycle operations
- Providers implement `IInfrastructureProvider` in addition to data processing
- Support commands: `{"command": "init", "config": {...}}` and `{"command": "destroy", "config": {...}}`

#### HCL Configuration Extension
```hcl
output {
  provider_path = "./azure-servicebus-provider"
  config {
    connection_string = "{{ env \"ASB_CONNECTION\" }}"
    infrastructure {
      pulumi_project = "./infrastructure/asb-queues"
      queue_naming_pattern = "{table_name}_events"
    }
  }
}
```

#### Implementation Details
- **Provider Autonomy**: Each provider manages its own infrastructure
- **Schema Awareness**: Input providers can inform output providers about required resources
- **Pulumi Integration**: Providers execute Pulumi programs for resource management
- **Language Agnostic**: Any language can implement infrastructure lifecycle via stdin/stdout
- **Composable**: Works with any provider combination

### TODO Items

#### CLI (Go)
- [ ] Add `init`, `destroy`, `plan`, `status` commands to CLI
- [ ] Implement lifecycle command routing in executor
- [ ] Add infrastructure result parsing and display
- [ ] Update HCL config parsing for infrastructure blocks

#### SDK (.NET)
- [ ] Add `IInfrastructureProvider` interface to abstractions
- [ ] Extend `StdioProviderHost` to handle lifecycle commands
- [ ] Add Pulumi runner utilities to SDK
- [ ] Create infrastructure result DTOs

#### Providers
- [ ] Update Azure Service Bus provider with infrastructure lifecycle
- [ ] Add MS SQL CDC provider with table discovery
- [ ] Create Pulumi templates for common scenarios (ASB queues, Kafka topics)

#### Documentation  
- [ ] Document infrastructure lifecycle workflow
- [ ] Add Pulumi integration examples
- [ ] Update provider development guide with infrastructure patterns

### Benefits Validated
✅ **stdin/stdout Flexibility**: Lifecycle commands work seamlessly with existing protocol
✅ **Language Agnostic**: Any language can implement infrastructure operations  
✅ **Provider Independence**: Each provider owns its infrastructure concerns
✅ **Composability**: Input schema can drive output infrastructure provisioning
✅ **Testability**: Infrastructure operations testable independently of data processing

### Notes
- This design maintains the Unix pipeline philosophy while adding infrastructure capabilities
- Providers remain simple binaries that handle both data and infrastructure via same stdin/stdout interface
- More flexible than gRPC approach would have been - no protocol constraints
- Aligns with "providers as independent executables" architecture

---

## Infrastructure Lifecycle Design - MS SQL CDC to Azure Service Bus (2025-09-18)

### Problem Statement
MS SQL CDC monitoring multiple tables needs dynamic Azure Service Bus queue creation:

```json
{
    "data": {
        "FirstName": "Ameer",
        "ID": "298", 
        "LastName": "Deen"
    },
    "metadata": {
        "LSN": "0000004c000028200003",
        "OperationID": 2,
        "OperationType": "Insert",
        "TableName": "Persons"
    }
}
```

### Design Decisions

#### ✅ **Embedded Pulumi Over External Terraform**
- **Pulumi embedded directly in provider binary**
- **Infrastructure code ships with provider in OCI image**
- **No external Terraform files to manage**
- **Infrastructure and code versions stay synchronized**

#### ✅ **Interface-Based Design**
```csharp
public interface IInfrastructureProvider : IProvider
{
    Task<InfrastructureResult> InitializeAsync(CancellationToken ct);
    Task<InfrastructureResult> DestroyAsync(CancellationToken ct);
    Task<InfrastructureResult> GetStatusAsync(CancellationToken ct);
    Task<InfrastructureResult> PlanAsync(CancellationToken ct);
    Task<string[]> DiscoverRequiredResourcesAsync(IEnumerable<string> tables, CancellationToken ct);
}

// Provider implements both data and infrastructure
public class AsbOutputProvider : InfrastructureProviderBase<AsbConfig>, IOutputProvider
{
    protected override PulumiStackBuilder ConfigureStack(AsbConfig config)
    {
        return new PulumiStackBuilder("asb-queues")
            .AddQueues(config.Tables.Select(t => $"{t}_cdc_events"))
            .WithProperties(new QueueProperties { MaxSizeMB = 1024, TTL = TimeSpan.FromDays(14) });
    }
}
```

#### ✅ **Task-Level CLI Commands**
```bash
dstream init mssql-to-asb      # Provision infrastructure for task
dstream plan mssql-to-asb      # Show what would be created/destroyed
dstream run mssql-to-asb       # Run the data pipeline
dstream status mssql-to-asb    # Show current infrastructure state
dstream destroy mssql-to-asb   # Clean up infrastructure for task
```

#### ✅ **Clean HCL Configuration**
```hcl
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
    provider_path = "./asb-output-provider"
    config {
      connection_string = "{{ env \"ASB_CONNECTION\" }}"
      # Infrastructure configuration embedded in provider code
    }
  }
}
```

### Benefits
- **☁️ Self-Contained**: Provider binary includes all infrastructure logic
- **📦 OCI Distribution**: Everything ships in one container image
- **🔄 Versioning**: Infrastructure and code versions stay synchronized
- **🧪 Testability**: Test infrastructure changes with same CI/CD pipeline
- **🔧 Provider Autonomy**: Each provider manages its own infrastructure concerns
- **🎯 Clean Separation**: Data processing and infrastructure via same interface

### Next Challenge: Verb Routing
**Problem**: How to pass lifecycle verbs (init/run/destroy/plan/status) to providers via stdin/stdout protocol?

**Current State**: Only supports single "run" operation
**Need**: Route different verbs to different provider methods based on CLI command
