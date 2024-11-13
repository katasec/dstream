// distributed_locker.go
package cdc

import (
	"context"
)

// DistributedLocker defines an interface for a distributed locking mechanism.
type DistributedLocker interface {
	// AcquireLock tries to acquire a lock and returns a lease ID if successful.
	AcquireLock(ctx context.Context) (string, error)

	// ReleaseLock releases the lock associated with the provided lease ID.
	ReleaseLock(ctx context.Context, leaseID string) error
}
