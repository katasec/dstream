package locking

import (
	"fmt"
)

// LockerFactory creates instances of DistributedLocker based on the configuration
type LockerFactory struct {
	//config           *config.Config
	connectionString string
	containerName    string
	configType       string
	//leaseDB *LeaseDBManager // Add LeaseDBManager for database operations
}

// NewLockerFactory initializes a new LockerFactory
// func NewLockerFactory(config *config.Config) *LockerFactory {
// 	return &LockerFactory{
// 		config: config,
// 		//leaseDB: leaseDB,
// 	}
// }

// NewLockerFactory initializes a new LockerFactory
func NewLockerFactory(configType string, connectionString string, containerName string) *LockerFactory {
	return &LockerFactory{
		// config: config,
		//leaseDB: leaseDB,
		containerName:    containerName,
		connectionString: connectionString,
		configType:       configType,
	}
}

// CreateLocker creates a DistributedLocker for the specified table
func (f *LockerFactory) CreateLocker(lockName string) (DistributedLocker, error) {
	switch f.configType {
	case "azure_blob":
		return NewBlobLocker(
			f.connectionString,
			f.containerName,
			lockName,
		)
	default:
		return nil, fmt.Errorf("unsupported lock type: %s", f.configType)
	}
}

// GetLockedTables checks if specific tables are locked
func (f *LockerFactory) GetLockedTables(tableNames []string) ([]string, error) {
	switch f.configType {
	case "azure_blob":
		// Create a temporary locker to check table locks
		tempLocker, err := NewBlobLocker(f.connectionString, f.containerName, "temp")
		if err != nil {
			return nil, fmt.Errorf("failed to create blob locker: %w", err)
		}
		return tempLocker.GetLockedTables(tableNames)
	default:
		return nil, fmt.Errorf("unsupported lock type: %s", f.configType)
	}
}
