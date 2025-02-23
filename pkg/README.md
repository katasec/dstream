# Public API Packages

This directory contains the public API packages that are intended to be used by external applications.

## Packages

### `cdc`
Contains interfaces and types for Change Data Capture (CDC) functionality:
- `interfaces.go`: Defines the core interfaces for table monitoring
- `types.go`: Contains CDC-related types like ChangeEvent and ChangeType

### `messaging`
Contains interfaces for message publishing:
- `interfaces.go`: Defines the Publisher interface for message publishing capabilities
