package cdc

import (
	"fmt"
	"time"

	"github.com/katasec/dstream/config"
)

var (
	defaultLockTTL = 120 * time.Second
)

// LockerFactory creates instances of DistributedLocker based on the configuration
type LockerFactory struct {
	config *config.Config
}

// NewLockerFactory initializes a new LockerFactory
func NewLockerFactory(config *config.Config) *LockerFactory {
	return &LockerFactory{config: config}
}

// CreateLocker creates a DistributedLocker for the specified table
func (f *LockerFactory) CreateLocker(lockName string) (DistributedLocker, error) {
	switch f.config.Locks.Type {
	case "azure_blob":
		return NewBlobLocker(
			f.config.Locks.ConnectionString,
			f.config.Locks.ContainerName,
			lockName,
			defaultLockTTL, // Default TTL for locks
		)
	default:
		return nil, fmt.Errorf("unsupported lock type: %s", f.config.Locks.Type)
	}
}
