// blob_locker.go
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
	defaultLockerContainerName = "cdc-locks"
)

// BlobLocker implements the DistributedLocker interface using Azure Blob Storage for distributed locking
type BlobLocker struct {
	client          *azblob.Client
	container       string
	lockTTL         time.Duration // Time-to-live for the lock
	blockblobClient *blockblob.Client
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

	// Ensure lockTTL is above the minimum threshold
	if lockTTL < time.Second*60 {
		lockTTL = time.Second * 60
	}

	return &BlobLocker{
		client:          client,
		container:       containerName,
		lockTTL:         lockTTL,
		blockblobClient: blockblobClient,
	}, nil
}

// AcquireLock tries to acquire a lock on the blob (lockName) by taking a lease and returns the lease ID if successful
func (bl *BlobLocker) AcquireLock(ctx context.Context, lockName string) (string, error) {

	// Ensure blob exists
	_, err := bl.blockblobClient.UploadBuffer(context.TODO(), []byte{}, &blockblob.UploadBufferOptions{})
	if err != nil && !strings.Contains(err.Error(), "BlobAlreadyExists") {
		return "", fmt.Errorf("failed to create or check blob %s: %w", lockName, err)
	}

	// Create a lease client for the blob
	blobLeaseClient, err := lease.NewBlobClient(bl.blockblobClient, &lease.BlobClientOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create lease client for blob %s: %w", lockName, err)
	}

	// Attempt to acquire a lease
	resp, err := blobLeaseClient.AcquireLease(ctx, int32(bl.lockTTL.Seconds()), nil)
	if err != nil {
		return "", fmt.Errorf("failed to acquire lock for blob %s: %w", lockName, err)
	}

	log.Printf("Lock acquired for blob %s with lease ID: %s", lockName, *resp.LeaseID)
	return *resp.LeaseID, nil
}

// ReleaseLock releases the lock associated with the provided lease ID for the specified blob (lockName)
func (bl *BlobLocker) ReleaseLock(ctx context.Context, lockName string) error {
	// Create a block blob client for the lockName
	blockblobClient, err := blockblob.NewClientFromConnectionString("", bl.container, lockName, nil)
	if err != nil {
		return fmt.Errorf("failed to create block blob client for %s: %w", lockName, err)
	}

	// Create a lease client for the blob
	blobLeaseClient, err := lease.NewBlobClient(blockblobClient, &lease.BlobClientOptions{})
	if err != nil {
		return fmt.Errorf("failed to create lease client for blob %s: %w", lockName, err)
	}

	// Attempt to release the lease
	_, err = blobLeaseClient.ReleaseLease(ctx, &lease.BlobReleaseOptions{})
	if err != nil {
		return fmt.Errorf("failed to release lock for blob %s: %w", lockName, err)
	}

	log.Printf("Lock released for blob %s", lockName)
	return nil
}
