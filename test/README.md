# Test Packages

This directory contains test-related code and utilities.

## Directory Structure

### `integration/`
Integration tests that test multiple components together:
- `testdb/`: Database-specific integration tests

### `mocks/`
Mock implementations of interfaces for testing:
- Mock publishers
- Mock monitors
- Other test doubles

## Running Tests

To run all tests including integration tests:
```bash
go test ./...
```

To run only unit tests (excluding integration):
```bash
go test $(go list ./... | grep -v /test/integration/)
```
