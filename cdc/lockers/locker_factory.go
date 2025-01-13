package lockers

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

// GetUnlockedTable Gets locked tables by the specified locked
func (f *LockerFactory) GetLockedTables() ([]string, error) {
	switch f.configType {
	case "azure_blob_db":
		lockedtables := []string{}
		return lockedtables, nil
	case "azure_blob":
		lockedtables := GetBlobLockerLockedTables(f.containerName, f.connectionString)
		return lockedtables, nil
	default:
		return nil, fmt.Errorf("unsupported lock type: %s", f.configType)
	}
}
