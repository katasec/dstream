# AGENTS.md - AI Agent Onboarding Guide for DStream

**Last Updated**: February 13, 2026  
**Purpose**: Architecture and integration guide for AI agents working on the DStream ecosystem

---

## 🎯 Quick Overview

### What is DStream?

**DStream is Terraform for data streaming.** It's a declarative, language-agnostic data pipeline orchestration system that:

- Uses **HCL configuration** (HashiCorp Configuration Language) to define data pipelines
- Implements a **provider model** where input and output providers can be written in any language
- Communicates via **stdin/stdout with JSON** (Unix pipeline philosophy)
- Distributes providers as **OCI artifacts** (like Docker images, but for binaries)
- Orchestrates **three-process architecture**: Input Provider → DStream CLI → Output Provider

### Key Concepts

- **HCL Config**: Declarative pipeline definitions (similar to Terraform)
- **Provider Model**: Pluggable input/output providers for different data sources/destinations
- **stdin/stdout Protocol**: Universal communication using JSON over pipes
- **OCI Distribution**: Cross-platform binaries distributed via container registries (GHCR)
- **Long-Running Services**: Providers are persistent processes that loop indefinitely

---

## 🏗️ Architecture Overview

### Three-Process Model

DStream uses a simple three-process orchestration model:

```
┌─────────────────┐         ┌──────────────┐         ┌──────────────────┐
│ Input Provider  │ stdout  │  DStream CLI │  stdin  │ Output Provider  │
│                 ├────────>│              ├────────>│                  │
│ (Generate data) │  JSON   │ (Orchestrate)│  JSON   │ (Consume data)   │
│                 │         │              │         │                  │
│ stderr → logs   │         │ stderr → logs│         │ stderr → logs    │
└─────────────────┘         └──────────────┘         └──────────────────┘
```

**Data Flow:**
1. **Input Provider** generates/polls data → writes JSON envelopes to stdout
2. **DStream CLI** reads from input provider's stdout → pipes to output provider's stdin
3. **Output Provider** reads JSON envelopes from stdin → processes/stores data

**Logging:**
- All providers write logs to **stderr** (never stdout)
- stdout is reserved exclusively for data flow

### Repository Ecosystem

| Repository | Role | Language | Link |
|------------|------|----------|------|
| `katasec/dstream` | CLI orchestrator | Go | https://github.com/katasec/dstream |
| `katasec/dstream-dotnet-sdk` | .NET SDK for providers | C# | https://github.com/katasec/dstream-dotnet-sdk |
| `katasec/dstream-ingester-mssql` | SQL Server CDC input provider | Go | https://github.com/katasec/dstream-ingester-mssql |
| `katasec/dstream-log-output-provider` | Log output provider | Go | https://github.com/katasec/dstream-log-output-provider |
| `katasec/dstream-counter-input-provider` | Counter/test input provider | C# | https://github.com/katasec/dstream-counter-input-provider |
| `katasec/dstream-console-output-provider` | Console output provider | C# | https://github.com/katasec/dstream-console-output-provider |

---

## 📋 Provider Contract

### Communication Protocol

Providers communicate with the DStream CLI using a simple stdin/stdout JSON protocol:

#### 1. Configuration (First Line from stdin)

The CLI sends a **command envelope** on the first line:

```json
{
  "command": "run",
  "config": {
    "db_connection_string": "server=localhost;database=TestDB;...",
    "poll_interval": "5s",
    "tables": ["Persons", "Cars"]
  }
}
```

- **`command`**: Lifecycle command (`run`, `init`, `plan`, `status`, `destroy`)
  - Input providers typically only support `run` (they're data generators)
  - Output providers support all lifecycle commands
- **`config`**: Provider-specific configuration from HCL

#### 2. Data Flow (Continuous JSON Envelopes)

After reading config, providers continuously read/write JSON envelopes:

```json
{
  "data": {
    "table_name": "Persons",
    "change_type": "insert",
    "id": 123,
    "name": "John Doe",
    "timestamp": "2025-09-28T20:00:00Z"
  },
  "metadata": {
    "TableName": "dbo.persons",
    "OperationType": "Insert",
    "LSN": "0000004c000028200003",
    "source": "mssql-cdc-provider"
  }
}
```

- **`data`**: Business payload (any JSON structure)
- **`metadata`**: Provider-specific context for routing, tracking, debugging

**Format**: JSON Lines (one envelope per line) - newline-delimited JSON

#### 3. Logging (All to stderr)

```go
// ✅ CORRECT - Log to stderr
log.Println("Processed batch of 100 records")  // Go default is stderr
Console.Error.WriteLine("Connection established");  // .NET

// ❌ WRONG - Never write logs to stdout
Console.WriteLine("Log message");  // This breaks data flow!
fmt.Println("Log message");  // This breaks data flow!
```

---

## 🔌 Provider Interfaces

### Critical Distinction: Long-Running Services

**⚠️ IMPORTANT**: Providers are **persistent processes**, not one-shot scripts!

- ✅ Read config **once** at startup
- ✅ Loop **indefinitely** generating/consuming data
- ✅ Graceful shutdown on SIGINT/SIGTERM
- ❌ **NOT** one-shot read-process-exit scripts

### .NET SDK Pattern

The `katasec/dstream-dotnet-sdk` provides a high-level abstraction:

#### Base Classes

```csharp
// Base class for all providers
public abstract class ProviderBase<TConfig>
{
    protected TConfig Config { get; private set; }
    protected IPluginContext Ctx { get; private set; }
    
    public void Initialize(TConfig config, IPluginContext ctx)
    {
        Config = config;
        Ctx = ctx;
    }
}
```

#### Input Provider Interface

```csharp
public interface IInputProvider : IProvider
{
    // Generate data continuously - this is an infinite stream!
    IAsyncEnumerable<Envelope> ReadAsync(
        IPluginContext ctx, 
        CancellationToken ct);
}
```

**Example: Counter Input Provider**

```csharp
public class CounterProvider : ProviderBase<CounterConfig>, IInputProvider
{
    public async IAsyncEnumerable<Envelope> ReadAsync(
        IPluginContext ctx, 
        [EnumeratorCancellation] CancellationToken ct)
    {
        int count = 0;
        
        // Loop indefinitely until cancelled
        while (!ct.IsCancellationRequested)
        {
            count++;
            
            yield return new Envelope
            {
                Data = new Dictionary<string, object>
                {
                    ["value"] = count,
                    ["timestamp"] = DateTime.UtcNow
                },
                Metadata = new Dictionary<string, object>
                {
                    ["source"] = "counter-provider"
                }
            };
            
            // Wait before generating next value
            await Task.Delay(Config.Interval, ct);
            
            // Optional: Stop after maxCount
            if (Config.MaxCount > 0 && count >= Config.MaxCount)
                break;
        }
    }
}

public class CounterConfig
{
    public int Interval { get; set; } = 1000;  // milliseconds
    public int MaxCount { get; set; } = 0;     // 0 = infinite
}
```

#### Output Provider Interface

```csharp
public interface IOutputProvider : IProvider
{
    // Process batches of data
    Task WriteAsync(
        IEnumerable<Envelope> batch, 
        IPluginContext ctx, 
        CancellationToken ct);
}
```

**Example: Console Output Provider**

```csharp
public class ConsoleProvider : ProviderBase<ConsoleConfig>, IOutputProvider
{
    public async Task WriteAsync(
        IEnumerable<Envelope> batch, 
        IPluginContext ctx, 
        CancellationToken ct)
    {
        foreach (var envelope in batch)
        {
            // Write to stderr (not stdout!)
            if (Config.OutputFormat == "simple")
            {
                Console.Error.WriteLine($"Message: {envelope.Data}");
            }
            else
            {
                Console.Error.WriteLine(JsonSerializer.Serialize(envelope));
            }
        }
        
        await Task.CompletedTask;
    }
}

public class ConsoleConfig
{
    public string OutputFormat { get; set; } = "simple";
}
```

#### Bootstrap

```csharp
// In Program.cs or Main method

// For input provider
await StdioProviderHost.RunInputProviderAsync<CounterProvider, CounterConfig>();

// For output provider
await StdioProviderHost.RunOutputProviderAsync<ConsoleProvider, ConsoleConfig>();
```

The SDK handles:
- Reading command envelope from stdin
- Parsing config
- Calling your `ReadAsync()` or `WriteAsync()` methods
- Writing JSON envelopes to stdout
- Signal handling and graceful shutdown

### Go Native Pattern

Go providers interact directly with stdin/stdout:

**Example: Input Provider (CDC/Polling Pattern)**

```go
package main

import (
    "bufio"
    "context"
    "encoding/json"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
)

type Config struct {
    DBConnectionString string   `json:"db_connection_string"`
    PollInterval       string   `json:"poll_interval"`
    Tables             []string `json:"tables"`
}

type CommandEnvelope struct {
    Command string          `json:"command"`
    Config  json.RawMessage `json:"config"`
}

type Envelope struct {
    Data     map[string]interface{} `json:"data"`
    Metadata map[string]interface{} `json:"metadata"`
}

func main() {
    // 1. Read command envelope from stdin (first line only)
    scanner := bufio.NewScanner(os.Stdin)
    if !scanner.Scan() {
        log.Fatal("Failed to read command envelope")
    }
    
    var cmdEnv CommandEnvelope
    if err := json.Unmarshal(scanner.Bytes(), &cmdEnv); err != nil {
        log.Fatalf("Failed to parse command envelope: %v", err)
    }
    
    var config Config
    if err := json.Unmarshal(cmdEnv.Config, &config); err != nil {
        log.Fatalf("Failed to parse config: %v", err)
    }
    
    log.Printf("Starting with config: %+v\n", config)  // stderr
    
    // 2. Setup context for graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Handle signals
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigChan
        log.Println("Shutdown signal received")
        cancel()
    }()
    
    // 3. Parse poll interval
    pollInterval, err := time.ParseDuration(config.PollInterval)
    if err != nil {
        pollInterval = 5 * time.Second
    }
    
    // 4. Main polling loop - runs indefinitely
    ticker := time.NewTicker(pollInterval)
    defer ticker.Stop()
    
    encoder := json.NewEncoder(os.Stdout)  // Write data to stdout
    
    for {
        select {
        case <-ctx.Done():
            log.Println("Shutting down gracefully")
            return
            
        case <-ticker.C:
            // Poll for changes
            changes := pollForChanges(config)
            
            // Write each change as JSON envelope to stdout
            for _, change := range changes {
                envelope := Envelope{
                    Data: change,
                    Metadata: map[string]interface{}{
                        "source":    "mssql-cdc-provider",
                        "timestamp": time.Now().UTC(),
                    },
                }
                
                if err := encoder.Encode(envelope); err != nil {
                    log.Printf("Error encoding envelope: %v", err)
                    return
                }
            }
            
            if len(changes) > 0 {
                log.Printf("Processed %d changes\n", len(changes))  // stderr
            }
        }
    }
}

func pollForChanges(config Config) []map[string]interface{} {
    // Your CDC/polling logic here
    // Query database, poll API, read files, etc.
    return []map[string]interface{}{}
}
```

**Example: Output Provider (Consumer Pattern)**

```go
package main

import (
    "bufio"
    "context"
    "encoding/json"
    "log"
    "os"
    "os/signal"
    "syscall"
)

type Config struct {
    LogLevel string `json:"logLevel"`
}

type CommandEnvelope struct {
    Command string          `json:"command"`
    Config  json.RawMessage `json:"config"`
}

type Envelope struct {
    Data     json.RawMessage        `json:"data"`
    Metadata map[string]interface{} `json:"metadata"`
}

func main() {
    // 1. Read command envelope from stdin (first line only)
    scanner := bufio.NewScanner(os.Stdin)
    if !scanner.Scan() {
        log.Fatal("Failed to read command envelope")
    }
    
    var cmdEnv CommandEnvelope
    if err := json.Unmarshal(scanner.Bytes(), &cmdEnv); err != nil {
        log.Fatalf("Failed to parse command envelope: %v", err)
    }
    
    var config Config
    if err := json.Unmarshal(cmdEnv.Config, &config); err != nil {
        log.Fatalf("Failed to parse config: %v", err)
    }
    
    log.Printf("Starting output provider with config: %+v\n", config)
    
    // 2. Setup context for graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigChan
        log.Println("Shutdown signal received")
        cancel()
    }()
    
    // 3. Read data envelopes from stdin (continuously after config line)
    lineCount := 0
    for scanner.Scan() {
        select {
        case <-ctx.Done():
            log.Println("Shutting down gracefully")
            return
        default:
        }
        
        line := scanner.Bytes()
        
        var envelope Envelope
        if err := json.Unmarshal(line, &envelope); err != nil {
            // If not valid JSON, just log the line as-is
            log.Printf("Received: %s\n", string(line))
            continue
        }
        
        // Process the envelope
        lineCount++
        log.Printf("[%d] Data: %s | Metadata: %v\n", 
            lineCount, string(envelope.Data), envelope.Metadata)
        
        // Your output logic here: write to database, send to API, etc.
    }
    
    if err := scanner.Err(); err != nil {
        log.Fatalf("Error reading stdin: %v", err)
    }
}
```

---

## 📦 Provider Distribution (OCI)

Providers are distributed as **OCI artifacts** (similar to Docker images, but for binaries):

### Build and Push

1. **Build cross-platform binaries**:
```bash
# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o out/linux-amd64/provider
GOOS=linux GOARCH=arm64 go build -o out/linux-arm64/provider
GOOS=darwin GOARCH=amd64 go build -o out/darwin-amd64/provider
GOOS=darwin GOARCH=arm64 go build -o out/darwin-arm64/provider
GOOS=windows GOARCH=amd64 go build -o out/windows-amd64/provider.exe
```

2. **Push to container registry** (e.g., GHCR):
```bash
# Using ORAS or similar OCI tools
oras push ghcr.io/katasec/my-provider:v1.0.0 \
  --artifact-type application/vnd.dstream.provider.v1 \
  ./out/linux-amd64/provider:application/vnd.dstream.binary.linux-amd64 \
  ./out/darwin-arm64/provider:application/vnd.dstream.binary.darwin-arm64
```

3. **Create provider metadata** (`provider.json`):
```json
{
  "name": "my-provider",
  "version": "1.0.0",
  "type": "input",
  "description": "Description of what this provider does",
  "config_schema": {
    "type": "object",
    "properties": {
      "api_key": {"type": "string"},
      "endpoint": {"type": "string"}
    },
    "required": ["api_key"]
  }
}
```

### Usage in HCL

```hcl
task "my-pipeline" {
  type = "providers"
  
  input {
    # Production - OCI reference
    provider_ref = "ghcr.io/katasec/my-provider:v1.0.0"
    
    config {
      api_key = "{{ env \"API_KEY\" }}"
    }
  }
  
  output {
    # Local development - file path
    provider_path = "../my-output-provider/out/provider"
    
    config {
      destination = "/data/output"
    }
  }
}
```

**How it works:**
- DStream CLI downloads provider binary on first use
- Caches locally in `~/.dstream/providers/`
- Selects correct binary for current OS/architecture
- Runs provider as subprocess

---

## 🧪 Testing Providers

### Local Testing Without CLI

Test providers independently using shell commands:

**Input Provider:**
```bash
# Build provider
go build -o my-input-provider

# Test with config
echo '{"command":"run","config":{"interval":1000,"maxCount":3}}' | \
  ./my-input-provider

# Expected output (to stdout):
# {"data":{"value":1},"metadata":{...}}
# {"data":{"value":2},"metadata":{...}}
# {"data":{"value":3},"metadata":{...}}
# (Logs go to stderr)
```

**Output Provider:**
```bash
# Build provider
go build -o my-output-provider

# Test with config and data
{
  echo '{"command":"run","config":{"logLevel":"info"}}'
  echo '{"data":{"id":1,"name":"Test"},"metadata":{}}'
  echo '{"data":{"id":2,"name":"Test2"},"metadata":{}}'
} | ./my-output-provider

# Expected: Logs to stderr showing processed data
```

**Pipe Test (Simulate DStream):**
```bash
# Test input → output pipeline
./my-input-provider | ./my-output-provider
```

### Integration Testing with DStream CLI

1. **Create test HCL config** (`test.hcl`):
```hcl
task "local-test" {
  type = "providers"
  
  input {
    provider_path = "./my-input-provider"
    config {
      interval = 1000
      maxCount = 5
    }
  }
  
  output {
    provider_path = "./my-output-provider"
    config {
      logLevel = "debug"
    }
  }
}
```

2. **Run with DStream**:
```bash
cd /path/to/dstream
go run . run local-test --log-level debug
```

3. **Verify**:
- Input provider generates data
- Output provider receives and processes data
- No errors in stderr
- Graceful shutdown on Ctrl+C

---

## 📚 Key Files Reference

### DStream Repository (`katasec/dstream`)

| File | Purpose | Link |
|------|---------|------|
| `WARP.md` | Project status and development context | [View](WARP.md) |
| `readme.md` | User documentation and quick start | [View](readme.md) |
| `AGENTS.md` | **This file** - Architecture guide | [View](AGENTS.md) |
| `dstream.hcl` | Example task configurations | [View](dstream.hcl) |
| `pkg/executor/providers.go` | Provider orchestration logic | [View](pkg/executor/providers.go) |
| `pkg/config/config.go` | HCL configuration parsing | [View](pkg/config/config.go) |
| `cmd/run.go` | CLI run command implementation | [View](cmd/run.go) |

### Provider Examples

| Provider | Language | Type | Repository |
|----------|----------|------|------------|
| MSSQL CDC | Go | Input | https://github.com/katasec/dstream-ingester-mssql |
| Log Output | Go | Output | https://github.com/katasec/dstream-log-output-provider |
| Counter | C# | Input | https://github.com/katasec/dstream-counter-input-provider |
| Console | C# | Output | https://github.com/katasec/dstream-console-output-provider |

### SDK Documentation

| SDK | Language | Repository |
|-----|----------|------------|
| .NET SDK | C# | https://github.com/katasec/dstream-dotnet-sdk |
| Go SDK | - | _(Native - no SDK needed, use stdin/stdout directly)_ |

---

## 💡 Common Development Patterns

### Pattern 1: Polling Input Provider (Go)

Full example for a provider that polls an external source:

```go
package main

import (
    "bufio"
    "context"
    "encoding/json"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
)

// Config represents provider-specific configuration
type Config struct {
    Endpoint     string   `json:"endpoint"`
    PollInterval string   `json:"poll_interval"`
    APIKey       string   `json:"api_key"`
}

// CommandEnvelope is the first line from stdin
type CommandEnvelope struct {
    Command string          `json:"command"`
    Config  json.RawMessage `json:"config"`
}

// Envelope is the data format for stdout
type Envelope struct {
    Data     map[string]interface{} `json:"data"`
    Metadata map[string]interface{} `json:"metadata"`
}

func main() {
    // Step 1: Read configuration from stdin (first line only)
    config := readConfig()
    
    // Step 2: Setup context and signal handling
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigChan
        log.Println("Received shutdown signal")
        cancel()
    }()
    
    // Step 3: Parse configuration
    pollInterval, _ := time.ParseDuration(config.PollInterval)
    if pollInterval == 0 {
        pollInterval = 5 * time.Second
    }
    
    log.Printf("Starting polling with interval: %v\n", pollInterval)
    
    // Step 4: Main polling loop
    ticker := time.NewTicker(pollInterval)
    defer ticker.Stop()
    
    encoder := json.NewEncoder(os.Stdout)
    
    for {
        select {
        case <-ctx.Done():
            log.Println("Shutting down")
            return
            
        case <-ticker.C:
            // Fetch data from external source
            items := fetchData(config)
            
            // Write each item as JSON envelope to stdout
            for _, item := range items {
                envelope := Envelope{
                    Data:     item,
                    Metadata: map[string]interface{}{
                        "source":    "polling-provider",
                        "timestamp": time.Now().UTC(),
                    },
                }
                
                if err := encoder.Encode(envelope); err != nil {
                    log.Printf("Error writing envelope: %v", err)
                    return
                }
            }
            
            log.Printf("Polled %d items\n", len(items))
        }
    }
}

func readConfig() Config {
    scanner := bufio.NewScanner(os.Stdin)
    if !scanner.Scan() {
        log.Fatal("Failed to read config")
    }
    
    var cmdEnv CommandEnvelope
    if err := json.Unmarshal(scanner.Bytes(), &cmdEnv); err != nil {
        log.Fatalf("Invalid command envelope: %v", err)
    }
    
    var config Config
    if err := json.Unmarshal(cmdEnv.Config, &config); err != nil {
        log.Fatalf("Invalid config: %v", err)
    }
    
    return config
}

func fetchData(config Config) []map[string]interface{} {
    // Your polling logic here
    // Make HTTP request, query database, read files, etc.
    return []map[string]interface{}{
        {"id": 1, "value": "example"},
    }
}
```

### Pattern 2: Stream Output Provider (Go)

Full example for a provider that writes to an external destination:

```go
package main

import (
    "bufio"
    "context"
    "encoding/json"
    "log"
    "os"
    "os/signal"
    "syscall"
)

type Config struct {
    Destination string `json:"destination"`
    BatchSize   int    `json:"batch_size"`
}

type CommandEnvelope struct {
    Command string          `json:"command"`
    Config  json.RawMessage `json:"config"`
}

type Envelope struct {
    Data     json.RawMessage        `json:"data"`
    Metadata map[string]interface{} `json:"metadata"`
}

func main() {
    // Step 1: Read configuration
    config := readConfig()
    
    // Step 2: Setup shutdown handling
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigChan
        log.Println("Received shutdown signal")
        cancel()
    }()
    
    log.Printf("Starting output to: %s\n", config.Destination)
    
    // Step 3: Initialize destination (connect to DB, open file, etc.)
    dest := initializeDestination(config)
    defer dest.Close()
    
    // Step 4: Read and process envelopes from stdin
    scanner := bufio.NewScanner(os.Stdin)
    batch := make([]Envelope, 0, config.BatchSize)
    
    for scanner.Scan() {
        select {
        case <-ctx.Done():
            // Flush remaining batch before shutdown
            if len(batch) > 0 {
                writeBatch(dest, batch)
            }
            log.Println("Shutdown complete")
            return
        default:
        }
        
        var envelope Envelope
        if err := json.Unmarshal(scanner.Bytes(), &envelope); err != nil {
            log.Printf("Invalid envelope: %v", err)
            continue
        }
        
        batch = append(batch, envelope)
        
        // Write batch when full
        if len(batch) >= config.BatchSize {
            writeBatch(dest, batch)
            batch = batch[:0]  // Reset batch
        }
    }
    
    // Write final batch
    if len(batch) > 0 {
        writeBatch(dest, batch)
    }
    
    if err := scanner.Err(); err != nil {
        log.Fatalf("Error reading stdin: %v", err)
    }
}

func readConfig() Config {
    scanner := bufio.NewScanner(os.Stdin)
    if !scanner.Scan() {
        log.Fatal("Failed to read config")
    }
    
    var cmdEnv CommandEnvelope
    if err := json.Unmarshal(scanner.Bytes(), &cmdEnv); err != nil {
        log.Fatalf("Invalid command envelope: %v", err)
    }
    
    var config Config
    if err := json.Unmarshal(cmdEnv.Config, &config); err != nil {
        log.Fatalf("Invalid config: %v", err)
    }
    
    // Set defaults
    if config.BatchSize == 0 {
        config.BatchSize = 100
    }
    
    return config
}

type Destination struct {
    // Your destination connection/handle
}

func (d *Destination) Close() error {
    // Cleanup
    return nil
}

func initializeDestination(config Config) *Destination {
    // Connect to database, open file, initialize API client, etc.
    return &Destination{}
}

func writeBatch(dest *Destination, batch []Envelope) {
    // Write batch to destination
    log.Printf("Writing batch of %d items\n", len(batch))
    
    // Your write logic here:
    // - Insert to database
    // - Send HTTP request
    // - Write to file
    // - Publish to message queue
}
```

### Pattern 3: Streaming Input Provider (C# .NET SDK)

```csharp
using Katasec.DStream.SDK;
using Katasec.DStream.Abstractions;
using System.Collections.Generic;
using System.Runtime.CompilerServices;
using System.Threading;
using System.Threading.Tasks;

public class StreamingInputProvider : ProviderBase<StreamConfig>, IInputProvider
{
    public async IAsyncEnumerable<Envelope> ReadAsync(
        IPluginContext ctx,
        [EnumeratorCancellation] CancellationToken ct)
    {
        // Initialize connection/stream
        using var stream = InitializeStream(Config);
        
        // Read continuously until cancelled
        while (!ct.IsCancellationRequested)
        {
            var items = await stream.ReadBatchAsync(ct);
            
            foreach (var item in items)
            {
                yield return new Envelope
                {
                    Data = item.ToDictionary(),
                    Metadata = new Dictionary<string, object>
                    {
                        ["source"] = "streaming-provider",
                        ["sequence"] = item.SequenceNumber
                    }
                };
            }
            
            // Optional: brief pause between batches
            if (items.Count == 0)
            {
                await Task.Delay(100, ct);
            }
        }
    }
    
    private IDataStream InitializeStream(StreamConfig config)
    {
        // Your stream initialization logic
        return new DataStream(config.Endpoint);
    }
}

public class StreamConfig
{
    public string Endpoint { get; set; }
    public string ApiKey { get; set; }
}

// In Program.cs
class Program
{
    static async Task Main(string[] args)
    {
        await StdioProviderHost.RunInputProviderAsync<StreamingInputProvider, StreamConfig>();
    }
}
```

---

## ❓ FAQ for AI Agents

### Q: How do providers communicate with the DStream CLI?

**A:** Via **stdin/stdout with JSON**:
- CLI writes config to provider's stdin (first line)
- Input providers write data envelopes to stdout
- Output providers read data envelopes from stdin
- All logs go to stderr (never stdout)

### Q: Are providers one-shot or long-running?

**A:** **Long-running services**! Critical distinction:
- Providers read config **once** at startup
- Then loop **indefinitely** generating/consuming data
- They're persistent processes, not one-shot scripts
- Graceful shutdown on SIGINT/SIGTERM

### Q: Where should logs go?

**A:** **Always stderr**, never stdout:
- ✅ `log.Println()` in Go (default is stderr)
- ✅ `Console.Error.WriteLine()` in C#
- ❌ Never `fmt.Println()` or `Console.WriteLine()` - breaks data flow

### Q: How to test providers locally?

**A:** Test with shell commands:
```bash
# Test input provider
echo '{"command":"run","config":{}}' | ./my-provider

# Test output provider
{
  echo '{"command":"run","config":{}}'
  echo '{"data":{"test":1},"metadata":{}}'
} | ./my-provider

# Test pipeline
./input-provider | ./output-provider
```

### Q: What's the data format?

**A:** **JSON Lines** (newline-delimited JSON):
```json
{"data":{...},"metadata":{...}}
{"data":{...},"metadata":{...}}
```
- One envelope per line
- `data`: Business payload (any JSON)
- `metadata`: Provider context

### Q: How are providers distributed?

**A:** As **OCI artifacts** (like Docker images):
- Build cross-platform binaries (Linux, macOS, Windows)
- Push to container registry (GHCR, Docker Hub)
- Reference in HCL: `provider_ref = "ghcr.io/org/provider:v1.0.0"`
- DStream downloads and caches automatically

### Q: Can I write providers in any language?

**A:** **Yes!** Any language that supports:
- Reading from stdin
- Writing to stdout
- JSON parsing/generation
- Examples: Go, C#, Python, Node.js, Rust, Java, etc.

### Q: What lifecycle commands do providers support?

**A:**
- **Input providers**: Typically only `run` (data generation)
- **Output providers**: All commands (`init`, `plan`, `status`, `destroy`, `run`)
- Command is in the first-line envelope: `{"command":"run","config":{...}}`

### Q: How does the CLI orchestrate providers?

**A:** Three-process model:
1. CLI spawns input provider subprocess → reads its stdout
2. CLI spawns output provider subprocess → writes to its stdin
3. CLI pipes data from input stdout → output stdin
4. Both providers log to stderr (visible in terminal)

### Q: What's the difference between `provider_ref` and `provider_path`?

**A:**
- **`provider_ref`**: OCI registry reference (production)
  - Example: `ghcr.io/katasec/my-provider:v1.0.0`
  - Auto-downloaded and cached
- **`provider_path`**: Local file path (development)
  - Example: `../my-provider/out/provider`
  - For local testing

### Q: How do I handle errors in providers?

**A:**
- Log errors to stderr: `log.Printf("Error: %v", err)`
- Exit with non-zero code for fatal errors
- CLI will detect provider exit and shutdown pipeline
- Use exponential backoff for transient errors

### Q: Can providers maintain state?

**A:** Yes, but carefully:
- Providers are **stateless between runs** (CLI can restart them)
- For persistence, use external storage (database, files, etc.)
- For checkpointing, store LSN/sequence numbers externally
- Don't rely on in-memory state surviving restarts

---

## 🚀 Next Steps for New Sessions

When starting a new session working on DStream:

### 1. Review Documentation
- [ ] Read `WARP.md` - Current project status and priorities
- [ ] Scan `readme.md` - User-facing documentation
- [ ] Review this file (`AGENTS.md`) - Architecture refresher

### 2. Understand Current State
- [ ] Check `dstream.hcl` - Example configurations
- [ ] Review latest commits - What changed recently?
- [ ] Read open issues/PRs - Current work in progress

### 3. Verify Environment
- [ ] Test CLI build: `cd /path/to/dstream && go build`
- [ ] Test example pipeline: `go run . run oci-counter-demo`
- [ ] Check provider repos are accessible

### 4. Locate Relevant Code
- [ ] **Provider orchestration**: `pkg/executor/providers.go`
- [ ] **HCL parsing**: `pkg/config/config.go`, `pkg/config/config_funcs.go`
- [ ] **CLI commands**: `cmd/*.go`
- [ ] **OCI fetching**: `pkg/orasfetch/fetch.go`

### 5. Understand Task Context
- [ ] What's the issue/feature?
- [ ] Which providers are involved?
- [ ] What HCL changes are needed?
- [ ] Are there related repositories to update?

### 6. Development Workflow
- [ ] Make minimal changes
- [ ] Test with local providers first (`provider_path`)
- [ ] Test with OCI providers (`provider_ref`)
- [ ] Verify logs go to stderr
- [ ] Test graceful shutdown (Ctrl+C)

### 7. Common Files to Modify

**For HCL features**:
- `pkg/config/config.go` - Data structures
- `pkg/config/config_funcs.go` - Parsing logic

**For provider orchestration**:
- `pkg/executor/providers.go` - Process management

**For CLI commands**:
- `cmd/*.go` - Command implementations

**For documentation**:
- `readme.md` - User docs
- `WARP.md` - Project status
- `AGENTS.md` - This file

---

## 📖 Additional Resources

### Repositories
- **Main CLI**: https://github.com/katasec/dstream
- **.NET SDK**: https://github.com/katasec/dstream-dotnet-sdk
- **Providers**: https://github.com/katasec?q=dstream

### Concepts
- **HCL**: https://github.com/hashicorp/hcl
- **OCI Artifacts**: https://github.com/opencontainers/artifacts
- **Unix Philosophy**: https://en.wikipedia.org/wiki/Unix_philosophy

### Similar Projects
- **Terraform**: Declarative infrastructure (inspiration)
- **Airbyte**: Data integration platform
- **Debezium**: Change data capture

---

## 🎓 Learning Path

### Beginner: Understanding DStream
1. Read the Quick Overview section
2. Run the counter-to-console example
3. Examine `dstream.hcl` configuration
4. Test with `--log-level debug` to see internals

### Intermediate: Creating Providers
1. Review Provider Contract section
2. Start with .NET SDK (easier) or Go native
3. Create a simple counter/echo provider
4. Test locally with shell pipes
5. Integrate with DStream CLI

### Advanced: Contributing
1. Understand three-process orchestration
2. Review `pkg/executor/providers.go`
3. Add HCL features or CLI commands
4. Build and distribute OCI providers
5. Contribute to ecosystem

---

**This guide is the single source of truth for DStream architecture. When in doubt, refer to the code examples and test them locally!**
