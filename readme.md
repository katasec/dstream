# DStream

**DStream is Terraform for data streaming.** Sources can be databases, APIs, files, queues—anything. Destinations can be databases, APIs, message brokers, data lakes—anywhere. 

Declare your data pipeline in HCL, run a single command, and DStream orchestrates everything.

## Quick Start (30 seconds)

Here's a real-world data pipeline that streams from a counter generator to console output:

**1. Create `dstream.hcl`:**
```hcl
task "my-pipeline" {
  type = "providers"
  
  input {
    provider_ref = "ghcr.io/writeameer/dstream-counter-input-provider:v0.3.0"
    config {
      interval = 1000    # Generate every 1 second
      maxCount = 5       # Stop after 5 messages
    }
  }
  
  output {
    provider_ref = "ghcr.io/writeameer/dstream-console-output-provider:v0.3.0"
    config {
      outputFormat = "simple"   # Clean output format
    }
  }
}
```

**2. Run your pipeline:**
```bash
go run . run my-pipeline
```

**3. See it work:**
```
[CounterInputProvider] Starting counter with interval=1000ms, max_count=5
Message #1: {"value":1,"timestamp":"2025-09-25T17:30:22.825803+00:00"}
Message #2: {"value":2,"timestamp":"2025-09-25T17:30:23.839113+00:00"}
Message #3: {"value":3,"timestamp":"2025-09-25T17:30:24.843170+00:00"}
Message #4: {"value":4,"timestamp":"2025-09-25T17:30:25.844992+00:00"}
Message #5: {"value":5,"timestamp":"2025-09-25T17:30:26.846957+00:00"}
✅ Task "my-pipeline" executed successfully
```

That's it! DStream automatically:
- 🚀 **Pulled providers** from the OCI registry (GHCR)
- 🔧 **Configured** both input and output providers
- 📡 **Streamed data** from counter to console in real-time
- 🛡️ **Handled** process lifecycle and graceful shutdown

---

## Why DStream?

### The Problem: Data Integration Complexity
- Moving data between systems requires custom code for each source/destination pair
- No standard way to compose, test, or deploy data pipelines
- Proprietary platforms lock you into specific languages, clouds, or vendors

### The Solution: Infrastructure as Code for Data
DStream applies **Terraform's philosophy** to data streaming:

- ✅ **Declarative**: Describe WHAT you want, not HOW to do it
- ✅ **Composable**: Mix and match any input with any output
- ✅ **Version-controlled**: Pipeline definitions live in Git
- ✅ **Cloud-agnostic**: Runs anywhere, supports any data source/destination
- ✅ **Language-agnostic**: Write providers in any language


## How It Works

DStream uses a **three-process orchestration model** inspired by Unix pipelines:

```
[Input Provider] ──stdin/stdout──> [DStream CLI] ──stdin/stdout──> [Output Provider]
```

### 1. Provider Distribution (OCI)
- **Providers are OCI artifacts** stored in container registries (GHCR, Docker Hub, etc.)
- **Cross-platform binaries** for Linux, macOS, Windows (x64/ARM64) 
- **Semantic versioning** with immutable, reproducible deployments
- **Automatic caching** - providers download once, cache locally

### 2. Pipeline Orchestration
- **DStream CLI** acts as the intelligent orchestrator
- **Launches provider processes** with proper configuration
- **Streams data** between providers using JSON over stdin/stdout
- **Handles lifecycle** - startup, monitoring, graceful shutdown

### 3. Universal Protocol
- **Language-agnostic** - providers can be written in any language 
- **Simple I/O** - JSON over stdin/stdout pipes (like Unix philosophy)
- **Easy testing** - test providers independently with shell commands
- **Zero dependencies** - no shared state or runtime coordination

---

## Real-World Examples

### Database CDC to Message Queue
```hcl
task "sql-to-kafka" {
  type = "providers"
  
  input {
    provider_ref = "ghcr.io/katasec/mssql-cdc-provider:v1.2.0"
    config {
      connection_string = "{{ env \"DATABASE_CONNECTION_STRING\" }}"
      tables = ["Orders", "Customers", "Inventory"]
      polling_interval = 1000
      batch_size = 100
    }
  }
  
  output {
    provider_ref = "ghcr.io/katasec/kafka-provider:v1.1.0"
    config {
      bootstrap_servers = "{{ env \"KAFKA_SERVERS\" }}"
      topic_prefix = "data_events"
      serialization = "json"
    }
  }
}
```

### REST API to Data Lake
```hcl
task "api-to-s3" {
  type = "providers"
  
  input {
    provider_ref = "ghcr.io/community/rest-api-provider:v2.0.0"
    config {
      endpoint = "https://api.example.com/events"
      auth_token = "{{ env \"API_TOKEN\" }}"
      poll_interval = 30000  # Every 30 seconds
    }
  }
  
  output {
    provider_ref = "ghcr.io/aws/s3-provider:v1.0.0"
    config {
      bucket = "my-data-lake"
      prefix = "events/{{ date \"2006-01-02\" }}/"
      format = "parquet"
    }
  }
}
```

### Local Development
```hcl
task "local-dev" {
  type = "providers"
  
  input {
    provider_path = "../my-custom-provider/out/my-provider"  # Local binary
    config {
      # Development configuration
    }
  }
  
  output {
    provider_ref = "ghcr.io/writeameer/dstream-console-output-provider:v0.3.0"
    config {
      outputFormat = "structured"
    }
  }
}
```

---

## How DStream Compares

| Category | DIY w/ Team + Tools | Enterprise (Striim, Fivetran) | OSS (Debezium / Kafka / Confluent) | DStream (OSS-first) |
|----------|---------------------|--------------------------------|-------------------------------------|----------------------|
| **Product License / Infra** | $4K–$8K/mo | $8K–$12K+/mo | $3K–$5K/mo (Confluent Cloud) or DIY infra | **$0 (always free)** |
| **Engineering Team (Dev, DevOps, Data Eng)** | $17K–$33K/mo (2–3 FTEs) | $8K–$17K/mo (still 1–2 FTEs for integration) | $12K–$20K/mo (1–2 FTEs for ops burden) | **$0 required** |
| **Complexity Overhead** | Medium–High | Low (managed, but lock-in) | High (Zookeeper, Kafka, backups) | **Low (Terraform-style config, pluggable providers)** |
| **Total Cost (TCO)** | $21K–$40K+/mo | $16K–$29K+/mo | $15K–$25K+/mo | **Free core; support starts at $2K/mo** |

### Why Teams Choose DStream:

- ✅ **Zero vendor lock-in** - Run anywhere, own your infrastructure
- ✅ **Terraform-familiar** - HCL config, declarative pipelines
- ✅ **Any language** - Write providers in Python, .NET, Rust, Go, Node.js
- ✅ **Any cloud** - AWS, Azure, GCP, or on-premises
- ✅ **Start free** - No licensing costs, no per-connector fees
- ✅ **Battle-tested** - Unix pipeline philosophy, process isolation

---

## Installation & Usage

**Prerequisites:** Go (latest version)

### 1. Get DStream
```bash
git clone https://github.com/katasec/dstream.git
cd dstream
go mod tidy
```

### 2. Create your pipeline
```hcl
# dstream.hcl
task "my-first-pipeline" {
  type = "providers"
  
  input {
    provider_ref = "ghcr.io/writeameer/dstream-counter-input-provider:v0.3.0"
    config {
      interval = 1000
      maxCount = 10
    }
  }
  
  output {
    provider_ref = "ghcr.io/writeameer/dstream-console-output-provider:v0.3.0"
    config {
      outputFormat = "simple"
    }
  }
}
```

### 3. Run your pipeline
```bash
go run . run my-first-pipeline
```

That's it! DStream will:
- 📥 **Pull** providers from OCI registry (cached locally)
- ⚡ **Launch** input and output provider processes
- 🔧 **Configure** each provider with your settings
- 🌊 **Stream** data from input to output in real-time
- ✨ **Handle** all process management and graceful shutdown

---

## Data Format

DStream uses a simple JSON envelope format for all data communication:

```json
{
  "data": {
    "id": 123,
    "name": "John Doe",
    "timestamp": "2025-09-25T17:30:22.825803+00:00"
  },
  "metadata": {
    "table": "users",
    "operation": "insert",
    "sequence": 42,
    "source": "mssql-cdc-provider"
  }
}
```

- **`data`**: Your business payload (any JSON structure)
- **`metadata`**: Provider-specific context for routing, tracking, and debugging
- **Format**: JSON Lines (one envelope per line) over stdin/stdout
- **Universal**: Works with any programming language

---

## Provider Ecosystem

### Available Providers

**Input Providers (Data Sources):**
- [Counter Input Provider](https://github.com/katasec/dstream-counter-input-provider) - Generate test data
- MS SQL CDC Provider (planned) - SQL Server Change Data Capture
- PostgreSQL CDC Provider (planned) - PostgreSQL logical replication
- REST API Provider (planned) - Poll REST endpoints
- Kafka Consumer Provider (planned) - Consume from Kafka topics
- File System Provider (planned) - Watch files and directories

**Output Providers (Data Destinations):**
- [Console Output Provider](https://github.com/katasec/dstream-console-output-provider) - Display to terminal
- Azure Service Bus Provider (planned) - Send to Azure Service Bus
- Kafka Producer Provider (planned) - Send to Kafka topics
- PostgreSQL Provider (planned) - Insert to PostgreSQL
- S3 Provider (planned) - Write to AWS S3
- File System Provider (planned) - Write to files

### Creating Your Own Providers

Providers can be written in **any language** that supports stdin/stdout:

#### Requirements
1. **Read configuration** from stdin (first line, JSON)
2. **For input providers**: Write JSON envelopes to stdout
3. **For output providers**: Read JSON envelopes from stdin  
4. **Write logs** to stderr (not stdout)
5. **Handle SIGTERM** for graceful shutdown

#### Development Options
- **[.NET SDK](https://github.com/katasec/dstream-dotnet-sdk)** - Full-featured SDK with abstractions
- **Python, Node.js, Rust, Java, etc.** - Direct stdin/stdout handling
- **Any language** that can process JSON and handle pipes

#### Distribution
- **Development**: Local binaries via `provider_path`
- **Production**: OCI artifacts via `provider_ref` (like Docker images)
- **Cross-platform**: Build for Linux/macOS/Windows, x64/ARM64

---

## Advanced Features

### Environment Variables
```hcl
config {
  connection_string = "{{ env \"DATABASE_URL\" }}"
  api_key = "{{ env \"API_KEY\" }}"
}
```

### Multiple Tasks
```bash
# List all tasks
go run . list

# Run specific task
go run . run my-pipeline

# Debug mode
go run . run my-pipeline --log-level debug
```

### Local Development vs Production
```hcl
# Local development
input {
  provider_path = "../my-provider/out/provider"  # Local binary
}

# Production deployment  
input {
  provider_ref = "ghcr.io/myorg/my-provider:v1.0.0"  # OCI registry
}
```

---

## Why DStream Works

✅ **Simple**: JSON over stdin/stdout - every language supports this
✅ **Reliable**: Process isolation prevents cascading failures
✅ **Testable**: Test each provider independently with shell commands
✅ **Scalable**: Providers are stateless, horizontally scalable processes
✅ **Universal**: Works on any OS, any language, any cloud

---

## Contributing

We welcome contributions:
- **New Providers** - Build connectors for your favorite systems
- **CLI Improvements** - Enhance the orchestration engine  
- **Documentation** - Help others understand the ecosystem

---

## License

MIT License - see LICENSE file for details.

---

> **DStream is "Terraform for data streaming"** - declarative, composable, and battle-tested.
> 
> Data pipelines should be as easy as `terraform apply` but for real-time streaming.
