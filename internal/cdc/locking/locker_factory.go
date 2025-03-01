package locking

import (
	"fmt"
	"strings"

	"github.com/katasec/dstream/internal/utils"
)

// LockerFactory creates instances of DistributedLocker based on the configuration
type LockerFactory struct {
	//config           *config.Config
	connectionString string
	containerName    string
	configType       string
	dbConnectionString string // Database connection string for server name extraction
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
func NewLockerFactory(configType string, connectionString string, containerName string, dbConnectionString string) *LockerFactory {
	return &LockerFactory{
		// config: config,
		//leaseDB: leaseDB,
		containerName:    containerName,
		connectionString: connectionString,
		configType:       configType,
		dbConnectionString: dbConnectionString,
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

// GetLockName returns the appropriate lock name for a given table name based on the locker type
func (f *LockerFactory) GetLockName(tableName string) string {
	switch f.configType {
	case "azure_blob":
		// For blob locker, use the blob-specific naming convention with server name subfolder
		if f.dbConnectionString != "" {
			// Extract server name from the database connection string
			serverName, err := utils.ExtractServerNameFromConnectionString(f.dbConnectionString)
			if err == nil && serverName != "" {
				// Create a hierarchical lock name with server name as subfolder
				return strings.ToLower(serverName) + "/" + GetBlobLockName(tableName)
			}
		}
		// Fall back to the default naming if we can't extract the server name
		return GetBlobLockName(tableName)
	default:
		// Default case, just return the table name as the lock name
		// This can be updated as new locker types are added
		return tableName
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
