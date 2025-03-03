package locking

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/lease"

	"github.com/katasec/dstream/internal/logging"
)

type BlobLocker struct {
	containerName string
	lockTTL       time.Duration
	lockName      string

	azblobClient    *azblob.Client
	blobLeaseClient *lease.BlobClient
	ctx             context.Context
}

func NewBlobLocker(connectionString, containerName, lockName string) (*BlobLocker, error) {

	// Create azblobClient and create container
	azblobClient, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure Blob client: %w", err)
	}
	_, err = azblobClient.CreateContainer(context.TODO(), containerName, nil)
	if err != nil && !strings.Contains(err.Error(), "ContainerAlreadyExists") {
		return nil, fmt.Errorf("failed to create or check container: %w", err)
	}

	// Create block blob client and upload empty blob
	blockblobClient, err := blockblob.NewClientFromConnectionString(connectionString, containerName, lockName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create block blob client: %w", err)
	}
	_, err = blockblobClient.UploadBuffer(context.TODO(), []byte{}, nil)
	if err != nil && !strings.Contains(err.Error(), "BlobAlreadyExists") && !strings.Contains(err.Error(), "412 There is currently a lease") {
		return nil, fmt.Errorf("failed to ensure blob exists: %w", err)
	}

	// create blobLease Client for srtuct
	blobLeaseClient, err := lease.NewBlobClient(blockblobClient, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create blob lease client: %w", err)
	}

	lockTTL := -1 * time.Second
	// if lockTTL < time.Second*60 {
	// 	lockTTL = time.Second * 60
	// }

	return &BlobLocker{
		containerName: containerName,
		lockTTL:       lockTTL,
		lockName:      lockName,

		azblobClient:    azblobClient,
		blobLeaseClient: blobLeaseClient,
		ctx:             context.TODO(),
	}, nil
}

// AcquireLock tries to acquire a lock on the blob and stores the lease ID
func (bl *BlobLocker) AcquireLock(ctx context.Context, lockName string) (string, error) {
	logger := logging.GetLogger()
	logger.Printf("Attempting to acquire lock for blob %s", bl.lockName)

	// Try to acquire lease
	resp, err := bl.blobLeaseClient.AcquireLease(bl.ctx, int32(bl.lockTTL.Seconds()), nil)
	if err != nil {
		// If there's already a lease, check its age
		if strings.Contains(err.Error(), "There is already a lease present") {
			// Get the blob's properties to check last modified time
			blobClient := bl.azblobClient.ServiceClient().NewContainerClient(bl.containerName).NewBlobClient(bl.lockName)
			props, err := blobClient.GetProperties(ctx, nil)
			if err != nil {
				return "", fmt.Errorf("failed to get blob properties for %s: %w", bl.lockName, err)
			}

			// Check if the lock is older than 2 minutes
			lastModified := props.LastModified
			lockAge := time.Since(*lastModified)
			logger.Printf("Lock on %s was last modified at: %v (%.2f minutes ago)", bl.lockName, lastModified.Format(time.RFC3339), lockAge.Minutes())

			if lockAge > 2*time.Minute {
				logger.Printf("Lock on %s is older than 2 minutes (last modified: %v). Breaking lease...", bl.lockName, lastModified.Format(time.RFC3339))

				// Break the lease
				_, err = bl.blobLeaseClient.BreakLease(ctx, nil)
				if err != nil {
					return "", fmt.Errorf("failed to break lease for %s: %w", bl.lockName, err)
				}

				// Wait a moment for the lease to be fully broken
				time.Sleep(time.Second)

				// Try to acquire the lease again
				resp, err = bl.blobLeaseClient.AcquireLease(ctx, int32(bl.lockTTL.Seconds()), nil)
				if err != nil {
					return "", fmt.Errorf("failed to acquire lease after breaking for %s: %w", bl.lockName, err)
				}
				logger.Printf("Successfully acquired lock after breaking old lease for %s", bl.lockName)
				return *resp.LeaseID, nil
			}

			logger.Printf("Table %s is already locked and the lock is still valid (%.2f minutes old, within 2 minute TTL). Skipping...", bl.lockName, lockAge.Minutes())
			return "", nil
		}
		return "", fmt.Errorf("failed to acquire lock for blob %s: %w", bl.lockName, err)
	}

	logger.Printf("Lock acquired for blob %s with Lease ID: %s", bl.lockName, *resp.LeaseID)
	return *resp.LeaseID, nil
}

func (bl *BlobLocker) RenewLock(ctx context.Context, lockName string) error {
	logger := logging.GetLogger()
	_, err := bl.blobLeaseClient.RenewLease(bl.ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to renew lock for blob %s: %w", lockName, err)
	}

	logger.Printf("Lock renewed for blob %s", lockName)
	return nil
}

// ReleaseLock releases the lock associated with the provided lease ID for the specified blob (lockName)
func (bl *BlobLocker) ReleaseLock(tx context.Context, lockName string, leaseID string) error {

	logger := logging.GetLogger()
	_, err := bl.blobLeaseClient.ReleaseLease(bl.ctx, &lease.BlobReleaseOptions{})
	if err != nil {
		return fmt.Errorf("failed to release lock for blob %s: %w", bl.lockName, err)
	} else {
		logger.Printf("Lock released successfully for blob %s", bl.lockName)
	}
	return nil
}

func (bl *BlobLocker) StartLockRenewal(ctx context.Context, lockName string) {
	logger := logging.GetLogger()
	logger.Printf("Starting lock renewal for blob %s", lockName)
	go func() {
		ticker := time.NewTicker(bl.lockTTL / 2)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := bl.RenewLock(bl.ctx, bl.lockName); err != nil {
					logger.Printf("Failed to renew lock for blob %s: %v", lockName, err)
				}
			case <-ctx.Done():
				logger.Printf("Stopping lock renewal for blob %s", lockName)
				return
			}
		}
	}()
}

// GetBlobLockName returns the lock name for a given table name using the blob locker naming convention
// This is a package-level function so it can be used by both BlobLocker and LockerFactory
func GetBlobLockName(tableName string) string {
	return tableName + ".lock"
}

// GetLockedTables checks if specific tables are locked
func (bl *BlobLocker) GetLockedTables(tableNames []string) ([]string, error) {
	lockedTables := []string{}
	containerClient := bl.azblobClient.ServiceClient().NewContainerClient(bl.containerName)

	// Check each table's lock status
	for _, tableName := range tableNames {
		lockName := GetBlobLockName(tableName)
		blobClient := containerClient.NewBlobClient(lockName)

		// Fetch the blob's properties
		resp, err := blobClient.GetProperties(context.TODO(), nil)
		if err != nil {
			// If blob doesn't exist, it's not locked
			if strings.Contains(err.Error(), "BlobNotFound") {
				continue
			}
			logger := logging.GetLogger()
			logger.Printf("Failed to get properties for blob %s: %v", lockName, err)
			continue
		}

		// Check the lease status and state
		leaseStatus := resp.LeaseStatus
		leaseState := resp.LeaseState
		lastModified := resp.LastModified
		lockAge := time.Since(*lastModified)

		if *leaseStatus == "locked" && *leaseState == "leased" {
			logger := logging.GetLogger()
			logger.Printf("Table %s is locked (last modified: %v, %.2f minutes ago)", 
				tableName, 
				lastModified.Format(time.RFC3339), 
				lockAge.Minutes())

			// Only consider the table locked if the lock is less than 2 minutes old
			if lockAge <= 2*time.Minute {
				logger.Printf(" - Lock is still valid (within 2 minute TTL)")
				lockedTables = append(lockedTables, lockName)
			} else {
				logger.Printf(" - Lock is stale (older than 2 minute TTL), will be broken when acquired")
			}
		}
	}

	return lockedTables, nil
}
