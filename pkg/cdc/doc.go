// Package cdc provides the public interfaces and types for Change Data Capture (CDC) functionality.
//
// The package defines core interfaces like TableMonitor for monitoring database changes,
// and types like ChangeEvent for representing database changes. These interfaces and types
// are meant to be used by external applications that want to integrate with the CDC system.
//
// Key Components:
//   - TableMonitor: Interface for monitoring database table changes
//   - TableMonitorFactory: Interface for creating table monitors
//   - ChangePublisher: Interface for publishing CDC changes
//   - ChangeEvent: Type representing a database change event
package cdc
