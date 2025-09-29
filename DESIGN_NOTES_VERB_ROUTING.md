# Provider Verb Routing Design for Infrastructure Lifecycle

## Problem Statement

We need to implement verb routing to support infrastructure lifecycle management (init/destroy/plan/status) through the existing stdin/stdout protocol. The current implementation only supports the "run" operation.

### Current Protocol (Data Flow Only)

```bash
# Current data flow protocol
echo '{...config JSON...}' | ./provider-binary
# Provider reads config from stdin, processes data, outputs to stdout
```

## Design Options

### Option A: Command Header in JSON Config

Extend the existing JSON configuration to include a command property:

```json
{
  "command": "init",  // New field: init, run, destroy, plan, status
  "config": {
    "connection_string": "...",
    "tables": ["Persons", "Orders"]
  }
}
```

**Pros:**
- Simple extension of existing protocol
- No changes to the CLI process launching mechanism
- Minimal changes to providers

**Cons:**
- Not immediately obvious from CLI code what operation is being performed
- Mixes control flow with configuration

### Option B: Environment Variable

Pass the command through an environment variable when launching the provider:

```bash
# CLI sets environment variable before launching provider
DSTREAM_COMMAND=init ./provider-binary
```

The provider reads both the environment variable and stdin configuration.

**Pros:**
- Clear separation between command and configuration
- Existing stdin/stdout flow unchanged

**Cons:**
- Provider needs to check environment variables and stdin
- Slightly less consistent with Unix pipeline philosophy

### Option C: Command Line Arguments

Pass the command as a command-line argument:

```bash
# CLI launches provider with command argument
./provider-binary --command init
```

**Pros:**
- Explicit and visible in process listing
- Standard CLI pattern

**Cons:**
- Providers need argument parsing
- Departs from pure stdin/stdout model

## Recommended Approach: Option A - Command Header in JSON Config

The command header in JSON config approach offers the best balance of:

1. Simplicity - extending existing JSON protocol
2. Compatibility - works with any language 
3. Consistency - maintains pure stdin/stdout model

### Implementation Details

#### CLI Executor Changes

```go
// ExecuteProviderTask enhanced to support commands
func ExecuteProviderTask(task *config.TaskBlock, command string) error {
  // Create config with command field
  inputConfig := map[string]interface{}{
    "command": command,           // Add command field
    "config": task.InputConfig,   // Existing config
  }
  
  // Serialize to JSON
  inputConfigJson, err := json.Marshal(inputConfig)
  
  // Send to provider
  fmt.Fprintln(inputStdin, string(inputConfigJson))
  
  // Similar for output provider
}
```

#### Provider Implementation (.NET)

```csharp
// StdioProviderHost.cs - Extended to handle commands
public static async Task RunAsync<TProvider, TConfig>(CancellationToken ct = default)
    where TProvider : ProviderBase<TConfig>, IProvider
    where TConfig : class
{
    // Read from stdin
    string json = await Console.In.ReadLineAsync();
    var commandEnvelope = JsonSerializer.Deserialize<CommandEnvelope<TConfig>>(json);
    
    // Create provider
    var provider = Activator.CreateInstance<TProvider>();
    provider.Configure(commandEnvelope.Config);
    
    // Route based on command
    switch (commandEnvelope.Command?.ToLowerInvariant())
    {
        case "init":
            if (provider is IInfrastructureProvider infraProvider)
                await HandleInitCommandAsync(infraProvider, ct);
            else
                throw new InvalidOperationException("Provider does not support infrastructure operations");
            break;
            
        case "destroy":
            if (provider is IInfrastructureProvider infraProvider)
                await HandleDestroyCommandAsync(infraProvider, ct);
            else
                throw new InvalidOperationException("Provider does not support infrastructure operations");
            break;
            
        case "run":
        case null:  // Default to run for backward compatibility
            await HandleRunCommandAsync(provider, ct);
            break;
            
        // Add cases for plan, status, etc.
            
        default:
            throw new InvalidOperationException($"Unknown command: {commandEnvelope.Command}");
    }
}

// Command envelope for deserialization
public class CommandEnvelope<TConfig>
{
    public string Command { get; set; }
    public TConfig Config { get; set; }
}
```

#### Output Provider Example

```csharp
public class AsbOutputProvider : InfrastructureProviderBase<AsbConfig>, IOutputProvider
{
    // Inherited from InfrastructureProviderBase:
    // - InitializeAsync() - Create queues
    // - DestroyAsync() - Delete queues
    // - PlanAsync() - Show what would be created/destroyed
    // - GetStatusAsync() - Show current state
    
    // IOutputProvider implementation for data flow
    public async Task WriteAsync(IEnumerable<Envelope> batch, IPluginContext ctx, CancellationToken ct)
    {
        foreach (var envelope in batch)
        {
            string tableName = envelope.Meta["TableName"]?.ToString();
            string queueName = $"{tableName}_cdc_events";
            
            // Ensure queue exists (idempotent operation)
            await EnsureQueueExistsAsync(queueName, ct);
            
            // Send message to queue
            await SendMessageAsync(queueName, envelope, ct);
        }
    }
}
```

## Compatibility Considerations

1. **Backward Compatibility**: Providers should default to "run" when no command is specified
2. **Error Handling**: Clear error messages for unsupported commands
3. **Interface Detection**: Check if provider implements `IInfrastructureProvider` before sending lifecycle commands

## Testing Strategy

1. **Unit Tests**: Verify command routing in StdioProviderHost
2. **Integration Tests**: Test end-to-end flow with mock providers
3. **Manual Tests**: Validate real providers with all lifecycle commands

## Next Steps

1. Update Go CLI executor to add command field to config JSON
2. Extend .NET SDK's StdioProviderHost to handle command routing
3. Implement command handling in ASB output provider
4. Add unit and integration tests for command routing
5. Update documentation