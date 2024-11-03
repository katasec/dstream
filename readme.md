# DStream

**DStream**  is a lightweight, proof-of-concept (POC) application designed to monitor Microsoft SQL Server tables enabled with Change Data Capture (CDC) for updates. When changes are detected, **DStream**  streams the data to an Azure Event Hub for further processing, analytics, or event-driven applications. As a single-binary application with no external dependencies, **DStream**  is easy to install, deploy, and containerize, making it highly suitable for cloud-native and scalable environments.

This project showcases the feasibility of capturing and streaming real-time data changes, making it ideal for applications that require up-to-date insights or need to respond to business events as they occur.

## Key Features

- **CDC Monitoring**: Tracks changes (inserts, updates, deletes) on MS SQL Server tables enabled with CDC.
- **Data Streaming**: Sends detected changes to an Azure Event Hub, enabling real-time data processing in downstream applications.
- **Flexible Configuration**: Easily configure target tables, polling intervals, and Event Hub connection settings. Features adaptive backoff for table monitoring, adjusting the polling frequency based on update rates to optimize performance and *reduce overhead on the monitored database server*.

## Use Cases

- **Event-Driven Architectures**: Use DStream to trigger downstream services when specific changes occur in the database.
- **Real-Time Analytics**: Stream data changes to analytics platforms for insights as events happen.
- **Data Replication**: Synchronize data across systems by streaming changes to other applications or services.
- **Business Monitoring**: Detect and respond to key business events in real time.

## Requirements

- **MS SQL Server** with CDC enabled on target tables
- **Azure Event Hub** for streaming data output
- **Go** (latest version recommended)

## Installation

1. **Clone the repository**:
   ```bash
   git clone https://github.com/yourusername/dstream.git
   cd dstream

2. **Configure environment variables**: Ensure that you have set up the necessary environment variables, such as the connection string for SQL Server and Azure Event Hub.

3. **Install dependencies:**
```bash
go mod tidy
```

## Configuration
In the config.hcl file, specify your MS SQL Server connection details, target tables, polling intervals, and Azure Event Hub settings.

Example `config.hcl`:

```hcl
db_type = "sqlserver"
db_connection_string = "sqlserver://<USERNAME>:<PASSWORD>@<SERVER_NAME>:1433?database=<DATABASE_NAME>"

output {
    type = "console"
}

tables {
    name = "Persons"
    poll_interval = "5s"
    max_poll_interval = "1m"
}

tables {
    name = "Cars"
    poll_interval = "10s"
    max_poll_interval = "2m"
}
```

## Usage

To start the application, run:

```bash
go run main.go
```

## Logging

DStream uses console logging by default. Logging levels and formats can be customized to suit your development or production needs.

## Future Enhancements

- **Multi-Database Support:** Extend monitoring capabilities to other databases (e.g., PostgreSQL, MySQL).
- **Output Sinks:** Support additional output types beyond Azure Event Hub (e.g., Kafka, AWS Kinesis, custom webhooks).
- **Dynamic Configuration:** Enable on-the-fly configuration updates for table and output settings.

## Contributing

Contributions are welcome! Please submit a pull request or create an issue if you encounter bugs or have suggestions for new features.

## License

This project is licensed under the MIT License. See the LICENSE file for details.