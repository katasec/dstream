
# DStream

**DStream** is a universal data streaming orchestration CLI that connects independent input and output providers using stdin/stdout communication. It supports both legacy Change Data Capture (CDC) workflows and modern provider-based architectures for streaming data from any source to any destination.

---

## Architecture Evolution

### 🔄 Two Execution Models

**1. Legacy Plugin Mode (`type = "plugin"`):**
- Single .NET plugin binary communicating via gRPC
- HashiCorp go-plugin protocol
- Backward compatibility maintained

**2. Modern Provider Mode (`type = "providers"`):**
- Independent input and output provider binaries
- Unix stdin/stdout communication
- Language-agnostic ecosystem
- Composable, testable, debuggable

### ✅ Unix Pipeline Philosophy
- **Simple I/O**: JSON over stdin/stdout pipes
- **Process Isolation**: Each provider runs independently
- **Universal Compatibility**: Works with any programming language
- **Easy Testing**: Test providers directly with shell commands
- **Operational Simplicity**: Standard Unix tooling and patterns

### ✅ Provider Independence
- **Standalone Binaries**: Each provider is a self-contained executable
- **Zero Dependencies**: No shared state or coordination required
- **Easy Development**: Focus on business logic, CLI handles orchestration
- **Fault Isolation**: Provider failures don't affect others

> **Modern DStream**: Think "Unix pipeline for data" - simple, composable, battle-tested.



## Modern Architecture: Provider Orchestration

The modern DStream architecture is built around a three-process orchestration model, inspired by Unix pipelines. DStream CLI acts as the orchestrator that launches, configures, and manages the communication between independent input and output provider processes.

### Three-Process Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Input Provider │    │   DStream CLI   │    │ Output Provider │
│   (Process 1)   │━━━▶│   (Process 2)   │━━━▶│   (Process 3)   │
│                 │    │  Orchestrator   │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
      stdout/stderr          stdin/stdout          stdout/stderr
```

### Process Communication Flow

1. **CLI launches both provider processes**
   - Starts input and output provider binaries as separate processes
   - Creates stdin/stdout/stderr pipes for inter-process communication
   - Each provider remains completely independent with no shared memory

2. **Configuration phase**
   - CLI sends JSON configuration to input provider via stdin
   - CLI sends JSON configuration to output provider via stdin
   - Each provider reads its configuration and initializes

3. **Data streaming phase**
   - Input provider generates data and writes JSON envelopes to stdout
   - CLI reads from input provider's stdout
   - CLI forwards data to output provider's stdin
   - Output provider processes data and writes results to its stdout (or destination)

4. **Logging and monitoring**
   - Both providers write logs to stderr (forwarded to CLI stderr)
   - CLI monitors both processes for errors or completion
   - CLI handles graceful shutdown of both processes on completion or interruption

### Detailed Data Flow Diagram

```
Input Provider Process:              CLI Process:                 Output Provider Process:
┌─────────────────────┐             ┌────────────────┐           ┌─────────────────────┐
│ 1. Read config from │◀────────────┤ Send config to │           │                     │
│    stdin (JSON)     │             │ input provider │           │                     │
├─────────────────────┤             ├────────────────┤           │                     │
│                     │             │                │           │                     │
│ 2. Generate data    │             │                │           │                     │
│    (business logic) │             │                │           │                     │
│                     │             │                │           │                     │
├─────────────────────┤             ├────────────────┤           ├─────────────────────┤
│                     │             │                │           │                     │
│ 3. Write JSON       │─────────────▶ Read from      │           │                     │
│    envelopes to     │             │ input stdout   │           │                     │
│    stdout           │             │                │           │                     │
│                     │             ├────────────────┤           │                     │
│                     │             │                │─────────────▶ 1. Read config    │
│                     │             │ Send config to │           │    from stdin      │
│                     │             │ output provider│           │    (JSON)          │
│                     │             │                │           │                     │
│                     │             ├────────────────┤           ├─────────────────────┤
│                     │             │                │           │                     │
│                     │             │ Forward data   │─────────────▶ 2. Read data      │
│                     │             │ to output      │           │    from stdin      │
│                     │             │ stdin          │           │    (JSON envelopes) │
│                     │             │                │           │                     │
├─────────────────────┤             ├────────────────┤           ├─────────────────────┤
│ 4. Write logs to    │─────────────▶ Forward stderr │─────────────▶ 3. Write logs to   │
│    stderr           │             │ to console     │           │    stderr          │
└─────────────────────┘             └────────────────┘           └─────────────────────┘
```

### Provider Communication Protocol

> 📄 **Implementation:** See [`pkg/executor/providers.go`](pkg/executor/providers.go) for complete orchestration logic

#### Configuration Protocol (first line via stdin)
```json
{"interval": 1000, "max_count": 10}
```

#### Data Envelope Protocol (subsequent lines over stdout/stdin)
```json
{"data": {"value": 42}, "metadata": {"seq": 1, "source": "counter"}}
```

#### Logging Protocol (stderr for diagnostics)
```
[CounterInputProvider] Starting service...
[ConsoleOutputProvider] Processed 1 messages
```

### CLI Orchestrator Purpose

The **DStream CLI** serves as the intelligent orchestrator in this three-process architecture. Its primary responsibilities include:

#### Process Management
- **Launch Control**: Starts input and output provider processes on demand
- **Resource Management**: Creates and manages stdin/stdout/stderr pipes between processes
- **Lifecycle Management**: Handles process startup, monitoring, and graceful shutdown
- **Error Recovery**: Detects provider failures and initiates cleanup procedures

#### Data Pipeline Orchestration
- **Configuration Distribution**: Parses HCL configuration and distributes provider-specific configs
- **Data Flow Control**: Acts as the data pump, reading from input provider and forwarding to output provider
- **Protocol Translation**: Ensures proper JSON envelope format between providers
- **Flow Monitoring**: Logs data flow statistics and handles streaming errors

#### Operational Management
- **Task Coordination**: Ensures both providers start in correct sequence with proper configuration
- **Signal Handling**: Manages graceful shutdown on system interrupts (Ctrl+C, SIGTERM)
- **Log Aggregation**: Collects and forwards stderr from both providers for centralized logging
- **Status Reporting**: Provides real-time feedback on pipeline health and completion status

#### Why This Design Works

This orchestrator pattern solves the fundamental challenge of connecting independent streaming processes:

- **No Direct Connection**: Input and output providers never communicate directly
- **Protocol Simplicity**: Each provider only needs to handle stdin/stdout JSON
- **Fault Tolerance**: CLI can restart failed providers without affecting the other
- **Composability**: Any input provider can work with any output provider through the CLI
- **Development Experience**: Providers can be developed and tested completely independently

### Key Benefits of Three-Process Architecture

1. **Process Isolation**: Complete isolation between providers prevents cascading failures
2. **Universal Language Support**: Any language that can read/write stdin/stdout works as a provider
3. **Simple Interface**: Providers only need to implement stdin/stdout JSON handling
4. **Independent Testing**: Each provider can be tested independently using shell commands
5. **Resource Control**: OS-level resource limits can be applied to each process separately
6. **Easy Debugging**: Standard input/output redirection works for troubleshooting
7. **Clean Shutdown**: Process signals (SIGTERM) enable graceful shutdown
8. **Zero Dependencies**: Providers require no shared libraries or runtime coordination
9. **Horizontal Scalability**: Multiple CLI instances can run different tasks simultaneously

### Testing the Architecture

> 📄 **Testing Reference:** Provider examples at [`../dstream-counter-input-provider/`](../dstream-counter-input-provider/) and [`../dstream-console-output-provider/`](../dstream-console-output-provider/)

You can validate each component of the three-process architecture independently:

#### Test Input Provider Alone
```bash
# Navigate to input provider directory
cd ~/progs/dstream/dstream-counter-input-provider

# Send configuration and see JSON envelopes output
echo '{"interval": 1000, "max_count": 3}' | bin/Release/net9.0/osx-x64/counter-input-provider
```

#### Test Output Provider Alone
```bash
# Navigate to output provider directory  
cd ~/progs/dstream/dstream-console-output-provider

# Send configuration first, then pipe sample data
{
  echo '{"outputFormat": "simple"}'
  echo '{"data": {"value": 42}, "metadata": {"seq": 1}}'
  echo '{"data": {"value": 43}, "metadata": {"seq": 2}}'
} | bin/Release/net9.0/osx-x64/console-output-provider
```

#### Test Complete Pipeline via CLI
```bash
# Navigate to CLI directory
cd ~/progs/dstream/dstream

# Run the full three-process orchestration
go run . run counter-to-console

# Or with debug logging to see orchestration details
go run . run counter-to-console --log-level debug
```

#### Manual Pipeline Simulation
```bash
# You can even simulate the CLI orchestration manually with shell pipes:
cd ~/progs/dstream

# Start input provider with config, pipe to output provider with its config
{
  echo '{"outputFormat": "simple"}'
  echo '{"interval": 1000, "max_count": 5}' | ../dstream-counter-input-provider/bin/Release/net9.0/osx-x64/counter-input-provider
} | ../dstream-console-output-provider/bin/Release/net9.0/osx-x64/console-output-provider
```

This testing approach demonstrates the **Unix pipeline philosophy** in action - each component works independently and can be composed together using standard shell tools.

## Key Features

### Ingestion
- **CDC Monitoring**: Tracks changes (inserts, updates, deletes) on MS SQL Server tables enabled with CDC
- **Reliable Offset Management**: Updates CDC offsets only after successful publish to ingest queue
- **Distributed Locking**: Uses Azure Blob Storage for distributed locking in multi-instance deployments
- **Adaptive Polling**: Features adaptive backoff for table monitoring based on change frequency
- **Automatic Topic Creation**: Creates topics and subscriptions for each monitored table

### Routing
- **Optimized Publishing**: Pre-creates and caches publishers at startup for better performance
- **Reliable Delivery**: Ensures messages are properly delivered to destination topics
- **Message Preservation**: Maintains original message properties during routing
- **Automatic Topic Management**: Creates topics and subscriptions as needed

### General
- **Flexible Configuration**: HCL-based configuration with environment variable support
- **Structured Logging**: Built-in structured logging with configurable levels
- **High Availability**: Supports running multiple instances for redundancy
- **Message Metadata**: Includes rich metadata for proper message routing and tracking

## Installation

1. **Clone the repository**:
   ```bash
   git clone https://github.com/katasec/dstream.git
   cd dstream
   ```

2. **Install dependencies**:
   ```bash
   go mod tidy
   ```

3. **Build the CLI**:
   ```bash
   go build -o dstream
   ```

4. **Create your task configuration** (`dstream.hcl`):
   ```hcl
   task "my-pipeline" {
     type = "providers"
     
     input {
       provider_path = "./my-input-provider"
       config {
         # Input provider configuration
       }
     }
     
     output {
       provider_path = "./my-output-provider"
       config {
         # Output provider configuration
       }
     }
   }
   ```

## Configuration

> 📄 **Configuration Parsing:** See [`pkg/config/`](pkg/config/) for HCL parsing and [`pkg/config/tasks.go`](pkg/config/tasks.go) for task definitions

DStream uses HCL for task configuration. Here's an example `dstream.hcl`:

### Modern Provider Tasks

```hcl
# Independent provider orchestration (recommended)
task "counter-to-console" {
  type = "providers"  # New provider orchestration mode
  
  input {
    provider_path = "../dstream-counter-input-provider/bin/Release/net9.0/osx-x64/counter-input-provider"
    config {
      interval = 1000   # Generate counter every 1 second
      max_count = 50    # Stop after 50 iterations
    }
  }
  
  output {
    provider_path = "../dstream-console-output-provider/bin/Release/net9.0/osx-x64/console-output-provider"
    config {
      outputFormat = "simple"  # Use simple output format
    }
  }
}

# Future: OCI container image providers
task "production-pipeline" {
  type = "providers"
  
  input {
    provider_ref = "ghcr.io/katasec/mssql-cdc-provider:v1.0.0"
    config {
      connection_string = "{{ env \"DATABASE_CONNECTION_STRING\" }}"
      tables = ["Orders", "Customers"]
    }
  }
  
  output {
    provider_ref = "ghcr.io/katasec/azure-servicebus-provider:v1.0.0"
    config {
      connection_string = "{{ env \"MESSAGING_CONNECTION_STRING\" }}"
      queue_name = "data-events"
    }
  }
}
```

### Legacy Plugin Tasks

```hcl
# Legacy single plugin mode (backward compatibility)
task "dotnet-counter-plugin" {
  type = "plugin"
  plugin_path = "../dstream-dotnet-sdk/samples/dstream-dotnet-test/out/dstream-dotnet-test"
   
  config {
    interval = 500  # Plugin-level configuration
  }
  
  input {
    provider = "null"
    config {
      interval = 1000
    }
  }
  
  output {
    provider = "console"
    config {
      format = "json"
    }
  }
}
```

## Usage

> 📄 **CLI Implementation:** See [`cmd/run.go`](cmd/run.go) for command handling and [`pkg/executor/executor.go`](pkg/executor/executor.go) for task routing

### Running Provider Tasks (Modern)

```bash
# Run a provider orchestration task
go run . run counter-to-console

# With debug logging
go run . run counter-to-console --log-level debug
```

The CLI will:
1. Parse the task configuration from `dstream.hcl`
2. Launch input and output provider processes
3. Send JSON configuration to each provider via stdin
4. Pipe data from input provider stdout to output provider stdin
5. Handle graceful shutdown and error recovery

### Running Plugin Tasks (Legacy)

```bash
# Run a legacy plugin task
go run . run dotnet-counter-plugin

# With debug logging for troubleshooting
go run . run dotnet-counter-plugin --log-level debug
```

### Task Management

```bash
# List all available tasks
go run . list

# Show task configuration (planned)
go run . show counter-to-console

# Validate configuration
go run . validate
```

## Data Envelope Format

> 📄 **Data Types:** See [`../dstream-dotnet-sdk/sdk/Katasec.DStream.Abstractions/Envelope.cs`](../dstream-dotnet-sdk/sdk/Katasec.DStream.Abstractions/Envelope.cs) for .NET envelope definition

DStream uses a standard JSON envelope format for provider communication:

### Counter Example
```json
{
  "data": {
    "value": 42,
    "timestamp": "2025-09-14T17:11:21.5590040+00:00"
  },
  "metadata": {
    "seq": 42,
    "interval_ms": 1000,
    "provider": "counter-input-provider"
  }
}
```

### CDC Example (Future)
```json
{
  "data": {
    "FirstName": "Diana",
    "ID": "180",
    "LastName": "Williams"
  },
  "metadata": {
    "table": "Persons",
    "operation": "Insert",
    "lsn": "0000003600000b200003",
    "timestamp": "2025-09-14T10:30:45Z"
  }
}
```

### Envelope Structure

- **`data`**: The actual payload (business data)
- **`metadata`**: Provider-specific metadata for tracking, routing, and debugging
- **Format**: One JSON envelope per line (JSON Lines format)
- **Encoding**: UTF-8 text over stdin/stdout

## Requirements

### For Provider Tasks
- **Go** (latest version) for the DStream CLI
- **Provider binaries** (any language that supports stdin/stdout)
- **HCL configuration** file (`dstream.hcl`)

### For Legacy Plugin Tasks
- **Go** (latest version) for the DStream CLI
- **.NET plugin binaries** with gRPC support
- **HCL configuration** file (`dstream.hcl`)

### Example Providers
- [Counter Input Provider (.NET)](https://github.com/katasec/dstream-counter-input-provider)
- [Console Output Provider (.NET)](https://github.com/katasec/dstream-console-output-provider)
- [DStream .NET SDK](https://github.com/katasec/dstream-dotnet-sdk)

## Provider Ecosystem

> 📄 **Provider Development:** See [`../dstream-dotnet-sdk/`](../dstream-dotnet-sdk/) for .NET SDK and provider templates

### Available Providers

**Input Providers:**
- [Counter Input Provider](https://github.com/katasec/dstream-counter-input-provider) - Generate test counter data
- SQL Server CDC Provider (planned) - SQL Server Change Data Capture
- PostgreSQL CDC Provider (planned) - PostgreSQL replication
- REST API Provider (planned) - Poll REST endpoints

**Output Providers:**
- [Console Output Provider](https://github.com/katasec/dstream-console-output-provider) - Display data to console
- Azure Service Bus Provider (planned) - Send to Azure Service Bus
- Kafka Provider (planned) - Send to Apache Kafka
- Database Provider (planned) - Insert to databases

### Creating Providers

Providers can be written in **any language** that supports stdin/stdout:

**Key Requirements:**
1. Read JSON configuration from stdin (first line)
2. For input providers: Write JSON envelopes to stdout
3. For output providers: Read JSON envelopes from stdin
4. Write logs/status to stderr
5. Handle graceful shutdown (SIGTERM)

**Example Provider Languages:**
- .NET (using [DStream .NET SDK](https://github.com/katasec/dstream-dotnet-sdk))
- Python, Node.js, Rust, Java, etc. (direct stdin/stdout handling)

### Provider Distribution

**Current:** Local binaries via `provider_path`
**Future:** OCI container images via `provider_ref`

```hcl
# Local development
input {
  provider_path = "./my-provider"
}

# Production deployment (planned)
input {
  provider_ref = "ghcr.io/myorg/my-provider:v1.0.0"
}
```


## Getting Started

### Quick Example

1. **Get the example providers**:
   ```bash
   # Clone and build counter input provider
   git clone https://github.com/katasec/dstream-counter-input-provider
   cd dstream-counter-input-provider
   /usr/local/share/dotnet/dotnet publish -c Release
   
   # Clone and build console output provider  
   git clone https://github.com/katasec/dstream-console-output-provider
   cd dstream-console-output-provider
   /usr/local/share/dotnet/dotnet publish -c Release
   ```

2. **Create `dstream.hcl`**:
   ```hcl
   task "demo" {
     type = "providers"
     
     input {
       provider_path = "../dstream-counter-input-provider/bin/Release/net9.0/osx-x64/counter-input-provider"
       config {
         interval = 1000
         max_count = 5
       }
     }
     
     output {
       provider_path = "../dstream-console-output-provider/bin/Release/net9.0/osx-x64/console-output-provider"
       config {
         outputFormat = "simple"
       }
     }
   }
   ```

3. **Run the pipeline**:
   ```bash
   go run . run demo
   ```

## Contributing

Contributions are welcome! This includes:

- **New Providers**: Create providers in any language
- **CLI Improvements**: Enhance the orchestration engine
- **Documentation**: Help others understand the ecosystem
- **Bug Reports**: Issues and suggestions

Please submit pull requests or create issues for discussions.

## License

This project is licensed under the MIT License. See the LICENSE file for details.

---

## Vision

> **DStream is "Unix pipelines for data streaming"** - simple, composable, language-agnostic, and battle-tested. 

We believe data streaming should be as easy as `cat file.txt | grep "error" | wc -l` but for real-time data pipelines.
