# DStream

**DStream** is a robust application designed to monitor Microsoft SQL Server tables enabled with Change Data Capture (CDC) for updates. When changes are detected, DStream streams the data to Azure Service Bus for further processing, analytics, or event-driven applications. As a single-binary application with minimal external dependencies, DStream is easy to install, deploy, and containerize, making it highly suitable for cloud-native and scalable environments.

## Key Features

- **CDC Monitoring**: Tracks changes (inserts, updates, deletes) on MS SQL Server tables enabled with CDC
- **Data Streaming**: Streams detected changes to Azure Service Bus, enabling real-time data processing
- **Distributed Locking**: Uses Azure Blob Storage for distributed locking, ensuring reliable operation in multi-instance deployments
- **Flexible Configuration**: HCL-based configuration with environment variable support for secure credential management
- **Structured Logging**: Built-in structured logging with configurable log levels
- **Adaptive Polling**: Features adaptive backoff for table monitoring, adjusting polling frequency based on update rates

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

    topic {
        name = "ingest-topic"
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

To start the application:

```bash
go run . server --log-level debug
```

## Architecture

DStream follows a modular architecture with clear separation of concerns:

### Components

1. **Ingester**
   - Monitors SQL Server tables for changes using CDC
   - Uses distributed locking to support multiple instances
   - Publishes changes to a message queue (e.g., Azure Service Bus)

2. **Processor**
   - Consumes changes from the message queue
   - Processes and transforms data as needed
   - Routes messages to appropriate destinations using Publishers

3. **Publishers**
   - Pluggable components that handle message delivery
   - Implementations available for:
     - Azure Service Bus
     - Azure Event Hubs
     - Console (for debugging)
   - Easy to add new implementations

### Data Flow
```
[SQL Server] --> [Ingester] --> [Queue] --> [Processor] --> [Publisher] --> [Destination]
    |              |             |            |              |
    |              |             |            |              |- Service Bus
    |              |             |            |              |- Event Hubs
    |              |             |            |              |- Console
    |              |             |            |
    |              |             |            |- Process Data
    |              |             |            |- Route Messages
    |              |             |
    |              |             |- Buffer Changes
    |              |             |- Decouple Components
    |              |
    |              |- Monitor CDC
    |              |- Distributed Locking
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

### Command Line Interface

DStream provides two main commands:

1. **Ingester**
   ```bash
   dstream ingester --config config.hcl
   ```
   Starts the ingester service that monitors CDC and publishes changes

2. **Processor**
   ```bash
   dstream processor --config config.hcl
   ```
   Starts the processor service that consumes and processes changes


## Contributing

Contributions are welcome! Please submit a pull request or create an issue if you encounter bugs or have suggestions for new features.

## License

This project is licensed under the MIT License. See the LICENSE file for details.