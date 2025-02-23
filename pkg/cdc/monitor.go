package cdc

import "context"

// TableMonitor defines the interface for monitoring database table changes
type TableMonitor interface {
	// MonitorTable starts monitoring the specified table for changes
	MonitorTable(ctx context.Context) error

	// GetTableName returns the name of the table being monitored
	GetTableName() string
}

// TableMonitorFactory creates table monitors for specific tables
type TableMonitorFactory interface {
	// CreateMonitor creates a new table monitor for the specified table
	CreateMonitor(tableName string) (TableMonitor, error)
}
