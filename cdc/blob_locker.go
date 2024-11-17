package cdc

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

var (
// defaultLockerContainerName = "locks"
)

// BlobLocker implements the DistributedLocker interface using Azure Blob Storage for distributed locking
type BlobLocker struct {
	client          *azblob.Client
	container       string
	lockTTL         time.Duration // Time-to-live for the lock
	blockblobClient *blockblob.Client
	blobLeaseClient *lease.BlobClient
	leaseID         string
}

// NewBlobLocker initializes a new BlobLocker
func NewBlobLocker(connectionString, containerName string, blobName string, lockTTL time.Duration) (*BlobLocker, error) {
	// Create blob client
	client, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure Blob client: %w", err)
	}

	// Ensure container exists
	_, err = client.CreateContainer(context.TODO(), containerName, nil)
	if err != nil && !strings.Contains(err.Error(), "The specified container already exists") {
		return nil, fmt.Errorf("failed to create or check container: %w", err)
	} else {
		log.Printf("Container %s, exists", containerName)
	}

	// Create blockblob client
	blockblobClient, err := blockblob.NewClientFromConnectionString(connectionString, containerName, blobName, nil)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Ensure blob exists
	_, err = blockblobClient.UploadBuffer(context.TODO(), []byte{}, &blockblob.UploadBufferOptions{})
	if err != nil && !strings.Contains(err.Error(), "BlobAlreadyExists") && !strings.Contains(err.Error(), "currently a lease on the blob") {
		log.Fatal(err.Error())
	}

	// Create a blob lease client
	blobLeaseClient, err := lease.NewBlobClient(blockblobClient, &lease.BlobClientOptions{})
	if err != nil {
		log.Fatal(err.Error())
	}

	// Ensure lockTTL is above the minimum threshold
	if lockTTL < time.Second*60 {
		lockTTL = time.Second * 60
	}

	return &BlobLocker{
		client:          client,
		container:       containerName,
		lockTTL:         lockTTL,
		blockblobClient: blockblobClient,
		blobLeaseClient: blobLeaseClient,
	}, nil
}

// AcquireLock tries to acquire a lock on the blob (lockName) by taking a lease and returns the lease ID if successful
func (bl *BlobLocker) AcquireLock(ctx context.Context, lockName string) (string, error) {

	// Attempt to acquire a lease
	resp, err := bl.blobLeaseClient.AcquireLease(ctx, int32(bl.lockTTL.Seconds()), nil)
	if err != nil {
		// Check if the error indicates an existing lease
		if strings.Contains(err.Error(), "LeaseIdMissing") {
			log.Printf("Table %s is already locked. Skipping monitoring for this table...", lockName)
			return "", nil // Indicate the lock is already held
		}
		return "", fmt.Errorf("failed to acquire lock for blob %s: %w", lockName, err)
	}
	bl.leaseID = *resp.LeaseID

	log.Printf("Lock acquired for blob %s with lease ID: %s", lockName, *resp.LeaseID)
	return bl.leaseID, nil
}

// ReleaseLock releases the lock associated with the provided lease ID for the specified blob (lockName)
func (bl *BlobLocker) ReleaseLock(ctx context.Context, lockName string) error {

	// Attempt to release the lease
	_, err := bl.blobLeaseClient.ReleaseLease(ctx, &lease.BlobReleaseOptions{})
	if err != nil {
		return fmt.Errorf("failed to release lock for blob %s: %w", lockName, err)
	}

	log.Printf("Lock released for blob %s", lockName)
	return nil
}

// ReleaseLock releases the lock associated with the provided lease ID for the specified blob (lockName)
func (bl *BlobLocker) RenewLock(ctx context.Context, lockName string) error {

	// Attempt to renewed the lease
	_, err := bl.blobLeaseClient.RenewLease(ctx, &lease.BlobRenewOptions{})
	if err != nil {
		return fmt.Errorf("failed to renewed lock for blob %s: %w", lockName, err)
	}

	log.Printf("Lock renewed for blob %s", lockName)
	return nil
}

// StartLockRenewal starts a background goroutine to renew the lock periodically.
func (bl *BlobLocker) StartLockRenewal(ctx context.Context, lockName string) {
	log.Println("Starting lock renewal background service for bloblocker....")
	go func() {
		ticker := time.NewTicker(bl.lockTTL / 2) // Renew halfway before TTL expires
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := bl.RenewLock(ctx, lockName); err != nil {
					log.Printf("Failed to renew lock for blob %s: %v", lockName, err)
				} else {
					log.Printf("Blob lock renewed for: %s", lockName)
				}
			case <-ctx.Done():
				log.Printf("Stopping lock renewal for blob %s", lockName)
				return
			}
		}
	}()
}
