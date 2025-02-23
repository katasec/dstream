# Internal Implementation Packages

This directory contains the internal implementation packages that are not intended to be used by external applications.

## Packages

### `cdc`
Contains CDC-specific implementations:
- `locking/`: Distributed locking mechanisms
- `service/`: Core CDC services
- `sqlserver/`: SQL Server specific implementations
- `utils/`: Shared utilities

### `config`
Configuration handling and validation.

### `db`
Database operations and table metadata management.

### `ingester`
Core ingestion logic for processing database changes.

### `logging`
Centralized logging setup and utilities.

### `messaging`
Message publishing implementations:
- Service Bus publisher
- Event Hub publisher
- Console publisher
