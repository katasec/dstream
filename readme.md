
# DStream

**DStream** is a robust, stateless Change Data Capture (CDC) streaming solution designed to capture changes from Microsoft SQL Server and reliably deliver them to downstream systems through Azure Service Bus or any other message queue provider. Its purpose is simple: **collect and forward** streaming data (currently CDC, but extensible to APIs, webhooks, or other streams) to any downstream system.

---

## Design Philosophy

### ✅ Stateless by Design
- No in-memory workflows.
- No complex orchestration.
- No recovery logic inside the service.
- State (offsets, LSNs, messages) is stored **externally** (in SQL or the queue).
- Crash-safe: restart, reschedule, or scale out with zero warmup or coordination.

### ✅ Durability & Reliability Offloaded
- dstream **offloads durability, retries, and ordering** to external queue systems (Azure Service Bus, AWS SQS, Kafka).
- Exactly-once guarantees at the batch level via atomic offset commits **after successful sends**.

### ✅ Simplicity = Resilience
- Reduces fault scenarios.
- Eliminates workflow recovery issues.
- Cuts operational overhead.
- Easier to debug and operate.

### ✅ Portability
- Works across any cloud, queue system, source (SQL CDC, APIs, webhooks), and destination (queues, services).

> The message queue is the orchestrator. The checkpoint is the resume point. The consumer defines the final behavior.



## Architecture

DStream operates in two main stages:

1. **Ingestion Stage (Ingester)**:
   - Monitors SQL Server tables enabled with CDC
   - Captures changes (inserts, updates, deletes)
   - Publishes changes to a central ingest queue
   - Updates CDC offsets only after successful queue publish
   - Uses distributed locking for high availability

2. **Routing Stage (Router)**:
   - Consumes messages from the ingest queue
   - Routes messages to their destination topics
   - Pre-creates publishers at startup for optimal performance
   - Ensures reliable delivery to downstream systems

This architecture provides several benefits:
- Reliable capture and delivery of changes
- Proper sequencing of messages
- High availability through distributed locking
- Optimized performance with connection pooling
- Clear separation of concerns between ingestion and routing

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

## Requirements

- **MS SQL Server** with CDC enabled on target tables
- **Azure Service Bus** for message streaming
- **Azure Blob Storage** for distributed locking
- **Go** (latest version recommended)

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

3. **Configure environment variables**:
   ```bash
   export DSTREAM_DB_CONNECTION_STRING="sqlserver://user:pass@localhost:1433?database=TestDB"
   export DSTREAM_INGEST_CONNECTION_STRING="your-azure-service-bus-connection-string"
   export DSTREAM_BLOB_CONNECTION_STRING="your-azure-blob-storage-connection-string"
   export DSTREAM_PUBLISHER_CONNECTION_STRING="your-azure-service-bus-connection-string"
   export DSTREAM_LOG_LEVEL="debug"  # Optional, defaults to info
   ```

## Configuration

DStream uses HCL for configuration. Here's an example `dstream.hcl`:

```hcl
ingester {
    db_type = "sqlserver"
    db_connection_string = "{{ env \"DSTREAM_DB_CONNECTION_STRING\" }}"

    poll_interval_defaults {
        poll_interval = "5s"
        max_poll_interval = "2m"
    }

    queue {
        type = "servicebus"
        name = "ingest-queue"
        connection_string = "{{ env \"DSTREAM_INGEST_CONNECTION_STRING\" }}"
    }

    locks {
        type = "azure_blob"
        connection_string = "{{ env \"DSTREAM_BLOB_CONNECTION_STRING\" }}"
        container_name = "locks"
    }

    tables = ["Persons"]

    tables_overrides {
        overrides {
            table_name = "Persons"
            poll_interval = "5s"
            max_poll_interval = "10m"
        }
    }
}

publisher {
    source {
        type = "azure_service_bus"
        connection_string = "{{ env \"DSTREAM_PUBLISHER_CONNECTION_STRING\" }}"
    }

    output {
        type = "azure_service_bus"
        connection_string = "{{ env \"DSTREAM_PUBLISHER_CONNECTION_STRING\" }}"
    }
}
```

## Usage

### Starting the Ingester

The ingester captures changes from SQL Server and publishes them to the ingest queue:

```bash
# Start with debug logging
go run . ingester --log-level debug

# Start with info logging (default)
go run . ingester
```

The ingester will:
1. Create topics for each monitored table
2. Create a 'sub1' subscription for each topic
3. Begin monitoring tables for changes
4. Publish changes to the ingest queue
5. Update CDC offsets after successful publish

### Starting the Router

The router consumes messages from the ingest queue and routes them to destination topics:

```bash
# Start with debug logging
go run . router --log-level debug

# Start with info logging (default)
go run . router
```

The router will:
1. Pre-create publishers for all configured tables
2. Begin consuming messages from the ingest queue
3. Route messages to their destination topics
4. Ensure reliable delivery with proper sequencing

## Message Format

DStream uses a consistent message format throughout the pipeline:

```json
{
    "data": {
        "FirstName": "Diana",
        "ID": "180",
        "LastName": "Williams"
    },
    "metadata": {
        "Destination": "server.database.table.events",
        "IngestQueue": "ingest-queue",
        "LSN": "0000003600000b200003",
        "OperationID": 2,
        "OperationType": "Insert",
        "TableName": "Persons"
    }
}
```

### Message Fields

#### Data Section
- Contains the actual change data
- Includes all columns from the monitored table
- Values are preserved in their original types

#### Metadata Section
- `Destination`: Fully qualified destination topic name
- `IngestQueue`: Name of the central ingest queue
- `LSN`: Log Sequence Number from SQL Server CDC
- `OperationID`: Type of change (1=delete, 2=insert, 3=update before, 4=update after)
- `OperationType`: Human-readable operation type
- `TableName`: Source table name

### Running in Production

For production deployments:

```bash
go run . server --log-level info
```

## Architecture

DStream follows a modular architecture with clear separation of concerns:

### Components

1. **CDC Monitor**
   - Monitors SQL Server tables for changes using CDC
   - Uses adaptive polling with configurable intervals
   - Tracks changes using LSN (Log Sequence Numbers)

2. **Publisher Adapter**
   - Wraps publishers with additional metadata
   - Adds destination routing information
   - Provides a unified interface for all publishers

3. **Publishers**
   - Pluggable components that handle message delivery
   - Implementations available for:
     - Azure Service Bus
     - Azure Event Hubs
     - Console (for debugging)
   - Easy to add new implementations via the Publisher interface

### Data Flow
```
[SQL Server] --> [CDC Monitor] --> [Publisher Adapter] --> [Publisher] --> [Destination]
    |              |                    |                    |
    |              |                    |                    |- Service Bus
    |              |                    |                    |- Event Hubs
    |              |                    |                    |- Console
    |              |                    |
    |              |                    |- Add Metadata
    |              |                    |- Route Information
    |              |
    |              |- Track LSN
    |              |- Adaptive Polling
    |
    |- CDC Enabled Tables
```

### Design Principles

1. **Modularity**
   - Clear separation between components
   - Pluggable publishers for different destinations
   - Easy to extend and maintain

2. **Reliability**
   - Distributed locking for multiple instances
   - Message queuing for reliable delivery
   - Graceful shutdown handling

3. **Observability**
   - Structured logging throughout
   - Configurable log levels
   - Clear error reporting

4. **Configuration**
   - HCL-based configuration
   - Environment variable support
   - Per-table configuration options

### Message Format

DStream publishes changes in a standardized JSON format:

```json
{
  "data": {
    "Field1": "Value1",
    "Field2": "Value2"
  },
  "metadata": {
    "Destination": "topic-or-queue-name",
    "LSN": "00000034000025c80003",
    "OperationID": 2,
    "OperationType": "Insert|Update|Delete",
    "TableName": "TableName"
  }
}
```

The metadata includes:
- Destination for routing
- LSN for tracking
- Operation type (Insert=2, Update=4, Delete=1)
- Source table name


## Contributing

Contributions are welcome! Please submit a pull request or create an issue if you encounter bugs or have suggestions for new features.

## License

This project is licensed under the MIT License. See the LICENSE file for details.

---

## Enhanced Features Summary

### ✅ Reliable Offset Management
- **CheckpointManager** tracks LSN offsets.
- Only commits after successful batch publishing to ensure **exactly-once delivery**.

### ✅ Adaptive Polling
- Uses **BackOffManager** for dynamic backoff based on activity and errors.

### ✅ BatchSizer
- Dynamically optimizes batch sizes considering message size limits and table characteristics.

### ✅ Distributed Locking
- Ensures single active table monitoring across instances.

### ✅ HCL Configuration
- Simple, declarative config via `dstream.hcl`.

### ✅ Plugin-Ready Architecture
- **TableMonitor** and planned **Publisher** interfaces make sources and sinks swappable.

### ✅ Extensible Support
- Ready for future sources (Postgres, APIs) and destinations (Kafka, Event Hub).

---

## The Goal
> Keep dstream lean, focused, and stateless—so it’s reliable, resilient, and boring (in the best way).

## Release