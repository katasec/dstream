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
	logger.Info("Attempting to acquire lock for blob", "lockName", bl.lockName)

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
			logger.Info("Lock was last modified", "lockName", bl.lockName, "lastModified", lastModified.Format(time.RFC3339), "ageMinutes", lockAge.Minutes())

			if lockAge > 2*time.Minute {
				logger.Info("Lock is older than 2 minutes, breaking lease", "lockName", bl.lockName, "lastModified", lastModified.Format(time.RFC3339))

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
				logger.Info("Successfully acquired lock after breaking old lease", "lockName", bl.lockName)
				return *resp.LeaseID, nil
			}

			logger.Info("Table is already locked and the lock is still valid", "lockName", bl.lockName, "ageMinutes", lockAge.Minutes(), "ttlMinutes", 2)
			return "", nil
		}
		return "", fmt.Errorf("failed to acquire lock for blob %s: %w", bl.lockName, err)
	}

	logger.Info("Lock acquired for blob", "lockName", bl.lockName, "leaseID", *resp.LeaseID)
	return *resp.LeaseID, nil
}

func (bl *BlobLocker) RenewLock(ctx context.Context, lockName string) error {
	logger := logging.GetLogger()
	_, err := bl.blobLeaseClient.RenewLease(bl.ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to renew lock for blob %s: %w", lockName, err)
	}

	logger.Info("Lock renewed for blob", "lockName", lockName)
	return nil
}

// ReleaseLock releases the lock associated with the provided lease ID for the specified blob (lockName)
func (bl *BlobLocker) ReleaseLock(tx context.Context, lockName string, leaseID string) error {

	logger := logging.GetLogger()
	_, err := bl.blobLeaseClient.ReleaseLease(bl.ctx, &lease.BlobReleaseOptions{})
	if err != nil {
		return fmt.Errorf("failed to release lock for blob %s: %w", bl.lockName, err)
	} else {
		logger.Info("Lock released successfully for blob", "lockName", bl.lockName)
	}
	return nil
}

func (bl *BlobLocker) StartLockRenewal(ctx context.Context, lockName string) {
	logger := logging.GetLogger()
	logger.Info("Starting lock renewal for blob", "lockName", lockName)
	go func() {
		ticker := time.NewTicker(bl.lockTTL / 2)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := bl.RenewLock(bl.ctx, bl.lockName); err != nil {
					logger.Error("Failed to renew lock for blob", "lockName", lockName, "error", err)
				}
			case <-ctx.Done():
				logger.Info("Stopping lock renewal for blob", "lockName", lockName)
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
			logger.Error("Failed to get properties for blob", "lockName", lockName, "error", err)
			continue
		}

		// Check the lease status and state
		leaseStatus := resp.LeaseStatus
		leaseState := resp.LeaseState
		lastModified := resp.LastModified
		lockAge := time.Since(*lastModified)

		if *leaseStatus == "locked" && *leaseState == "leased" {
			logger := logging.GetLogger()
			logger.Info("Table is locked", 
				"tableName", tableName, 
				"lastModified", lastModified.Format(time.RFC3339), 
				"ageMinutes", lockAge.Minutes())

			// Only consider the table locked if the lock is less than 2 minutes old
			if lockAge <= 2*time.Minute {
				logger.Info("Lock is still valid", "ttlMinutes", 2)
				lockedTables = append(lockedTables, lockName)
			} else {
				logger.Info("Lock is stale, will be broken when acquired", "ttlMinutes", 2)
			}
		}
	}

	return lockedTables, nil
}
