package lockers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/lease"
)

type BlobLockerDb struct {
	client          *azblob.Client
	container       string
	lockTTL         time.Duration
	blockblobClient *blockblob.Client
	blobLeaseClient *lease.BlobClient
	leaseDB         *LeaseDBManager
}

func NewBlobLockerDb(connectionString, containerName, lockName string, lockTTL time.Duration, leaseDB *LeaseDBManager) (*BlobLockerDb, error) {
	client, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure Blob client: %w", err)
	}

	_, err = client.CreateContainer(context.TODO(), containerName, nil)
	if err != nil && !strings.Contains(err.Error(), "ContainerAlreadyExists") {
		return nil, fmt.Errorf("failed to create or check container: %w", err)
	}

	blockblobClient, err := blockblob.NewClientFromConnectionString(connectionString, containerName, lockName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create block blob client: %w", err)
	}

	_, err = blockblobClient.UploadBuffer(context.TODO(), []byte{}, nil)
	if err != nil && !strings.Contains(err.Error(), "BlobAlreadyExists") && !strings.Contains(err.Error(), "412 There is currently a lease") {
		return nil, fmt.Errorf("failed to ensure blob exists: %w", err)
	}

	blobLeaseClient, err := lease.NewBlobClient(blockblobClient, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create blob lease client: %w", err)
	}

	if lockTTL < time.Second*60 {
		lockTTL = time.Second * 60
	}

	return &BlobLockerDb{
		client:          client,
		container:       containerName,
		lockTTL:         lockTTL,
		blockblobClient: blockblobClient,
		blobLeaseClient: blobLeaseClient,
		leaseDB:         leaseDB,
	}, nil
}

// AcquireLock tries to acquire a lock on the blob and stores the lease ID
func (bl *BlobLockerDb) AcquireLock(ctx context.Context, lockName string) (string, error) {
	ttl := int32(bl.lockTTL.Seconds())
	log.Printf("Attempting to acquire lock for blob %s", lockName)

	resp, err := bl.blobLeaseClient.AcquireLease(ctx, ttl, nil)
	if err != nil {
		if strings.Contains(err.Error(), "There is already a lease present") {
			log.Printf("Table %s is already locked. Skipping...", lockName)
			return "", nil
		}
		return "", fmt.Errorf("failed to acquire lock for blob %s: %w", lockName, err)
	}

	leaseID := *resp.LeaseID
	log.Printf("Lock acquired for blob %s with Lease ID: %s", lockName, leaseID)

	// Store Lease ID in database
	err = bl.leaseDB.StoreLeaseID(lockName, leaseID)
	if err != nil {
		return "", fmt.Errorf("failed to persist lease ID for lock %s: %w", lockName, err)
	}

	return leaseID, nil
}

func (bl *BlobLockerDb) RenewLock(ctx context.Context, lockName string) error {
	_, err := bl.blobLeaseClient.RenewLease(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to renew lock for blob %s: %w", lockName, err)
	}

	log.Printf("Lock renewed for blob %s", lockName)
	return nil
}

func (bl *BlobLockerDb) StartLockRenewal(ctx context.Context, lockName string) {
	log.Printf("Starting lock renewal for blob %s", lockName)
	go func() {
		ticker := time.NewTicker(bl.lockTTL / 2)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := bl.RenewLock(ctx, lockName); err != nil {
					log.Printf("Failed to renew lock for blob %s: %v", lockName, err)
				}
			case <-ctx.Done():
				log.Printf("Stopping lock renewal for blob %s", lockName)
				return
			}
		}
	}()
}

// ReleaseLock releases the lock associated with the provided lease ID for the specified blob (lockName)
func (bl *BlobLockerDb) ReleaseLock(ctx context.Context, lockName string, leaseID string) error {
	// Use the provided leaseID directly to release the lock
	if leaseID == "" {
		return fmt.Errorf("lease ID cannot be empty for lock %s", lockName)
	}

	_, err := bl.blobLeaseClient.ReleaseLease(ctx, &lease.BlobReleaseOptions{})
	if err != nil {
		return fmt.Errorf("failed to release lock for blob %s: %w", lockName, err)
	}

	// Remove the lease ID from the database
	err = bl.leaseDB.DeleteLeaseID(lockName)
	if err != nil {
		return fmt.Errorf("failed to delete lease ID for lock %s: %w", lockName, err)
	}

	log.Printf("Lock released for blob %s with Lease ID: %s", lockName, leaseID)
	return nil
}
