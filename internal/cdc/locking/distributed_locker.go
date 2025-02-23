// distributed_locker.go
package locking

import (
	"context"
)

// DistributedLocker defines an interface for a distributed locking mechanism.
type DistributedLocker interface {
	// AcquireLock tries to acquire a lock for the given lockName and returns a lease ID if successful.
	AcquireLock(ctx context.Context, lockName string) (string, error)

	// ReleaseLock releases the lock associated with the provided lease ID for the given lockName.
	//ReleaseLock(ctx context.Context, lockName string) error
	ReleaseLock(ctx context.Context, lockName string, leaseID string) error

	// ReleaseLock releases the lock associated with the provided lease ID for the given lockName.
	RenewLock(ctx context.Context, lockName string) error

	// StartLockRenewal starts a background process to renew the lock periodically.
	StartLockRenewal(ctx context.Context, lockName string)

	// GetLockedTables checks if specific tables are locked and returns a list of locked table names.
	GetLockedTables(tableNames []string) ([]string, error)
}
