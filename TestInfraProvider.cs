using System.Text.Json;
using Katasec.DStream.Abstractions;
using Katasec.DStream.SDK.Core;

namespace TestInfraProvider;

// Configuration class for the test provider
public class TestConfig
{
    public string? Name { get; set; } = "test-provider";
    public string[]? Resources { get; set; } = Array.Empty<string>();
}

// Test infrastructure provider that demonstrates the command routing
public class TestInfraProvider : InfrastructureProviderBase<TestConfig>, IOutputProvider
{
    // Infrastructure lifecycle methods
    protected override async Task<string[]> OnInitializeInfrastructureAsync(CancellationToken ct)
    {
        await Console.Error.WriteLineAsync($"[{nameof(TestInfraProvider)}] Creating test resources...");
        
        var resources = new[]
        {
            $"{Config.Name}_queue_1",
            $"{Config.Name}_queue_2",
            $"{Config.Name}_queue_3"
        };
        
        // Simulate resource creation
        await Task.Delay(1000, ct);
        
        await Console.Error.WriteLineAsync($"[{nameof(TestInfraProvider)}] Created {resources.Length} test resources");
        
        return resources;
    }
    
    protected override async Task<string[]> OnDestroyInfrastructureAsync(CancellationToken ct)
    {
        await Console.Error.WriteLineAsync($"[{nameof(TestInfraProvider)}] Destroying test resources...");
        
        var resources = new[]
        {
            $"{Config.Name}_queue_1",
            $"{Config.Name}_queue_2", 
            $"{Config.Name}_queue_3"
        };
        
        // Simulate resource destruction
        await Task.Delay(1000, ct);
        
        await Console.Error.WriteLineAsync($"[{nameof(TestInfraProvider)}] Destroyed {resources.Length} test resources");
        
        return resources;
    }
    
    protected override async Task<(string[] resources, Dictionary<string, object?>? metadata)> OnGetInfrastructureStatusAsync(CancellationToken ct)
    {
        await Console.Error.WriteLineAsync($"[{nameof(TestInfraProvider)}] Checking resource status...");
        
        var resources = new[]
        {
            $"{Config.Name}_queue_1",
            $"{Config.Name}_queue_2",
            $"{Config.Name}_queue_3"
        };
        
        var metadata = new Dictionary<string, object?>
        {
            ["provider_type"] = "test",
            ["health"] = "healthy",
            ["last_check"] = DateTime.UtcNow
        };
        
        await Task.Delay(500, ct);
        
        return (resources, metadata);
    }
    
    protected override async Task<(string[] resources, Dictionary<string, object?>? changes)> OnPlanInfrastructureChangesAsync(CancellationToken ct)
    {
        await Console.Error.WriteLineAsync($"[{nameof(TestInfraProvider)}] Planning infrastructure changes...");
        
        var resources = new[]
        {
            $"{Config.Name}_queue_1",
            $"{Config.Name}_queue_2",
            $"{Config.Name}_queue_3"
        };
        
        var changes = new Dictionary<string, object?>
        {
            ["action"] = "create",
            ["resource_count"] = resources.Length,
            ["estimated_cost"] = "$0.00"
        };
        
        await Task.Delay(500, ct);
        
        return (resources, changes);
    }
    
    // Output provider implementation (for run command)
    public async Task WriteAsync(IEnumerable<Envelope> batch, IPluginContext ctx, CancellationToken ct)
    {
        await Console.Error.WriteLineAsync($"[{nameof(TestInfraProvider)}] Processing batch of {batch.Count()} envelopes");
        
        foreach (var envelope in batch)
        {
            await Console.Error.WriteLineAsync($"[{nameof(TestInfraProvider)}] Processed envelope with payload: {JsonSerializer.Serialize(envelope.Payload)}");
        }
    }
}

// Main entry point
public class Program
{
    public static async Task Main(string[] args)
    {
        await StdioProviderHost.RunProviderWithCommandAsync<TestInfraProvider, TestConfig>();
    }
}