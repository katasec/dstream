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
	blobName        string
	lockTTL         time.Duration // Time-to-live for the lock
	blobLeaseClient *lease.BlobClient
}

// NewBlobLocker initializes a new BlobLocker
func NewBlobLocker(connectionString, containerName, blobName string, lockTTL time.Duration) (*BlobLocker, error) {

	// Create blob client
	client, err := azblob.NewClientFromConnectionString(connectionString, nil)
	handleError(err)

	// Ensure container exists
	_, err = client.CreateContainer(context.TODO(), containerName, nil)
	alreadyExists := strings.Contains(err.Error(), "The specified container already exists")
	if err != nil && !alreadyExists {
		handleError(fmt.Errorf("failed to create or check container: %w", err))
	}

	// Ensure blob exists
	blockblobClient, err := blockblob.NewClientFromConnectionString(connectionString, containerName, blobName, &blockblob.ClientOptions{})
	handleError(err)
	_, err = blockblobClient.UploadBuffer(context.TODO(), []byte{}, &blockblob.UploadBufferOptions{})
	handleError(err)

	// Create Lease client
	blobLeaseClient, err := lease.NewBlobClient(blockblobClient, &lease.BlobClientOptions{})
	handleError(err)

	// lockTTL min value
	if lockTTL < 60 {
		lockTTL = 60
	}

	return &BlobLocker{
		client:          client,
		container:       defaultLockerContainerName,
		blobName:        blobName,
		lockTTL:         lockTTL,
		blobLeaseClient: blobLeaseClient,
	}, nil
}

// AcquireLock tries to acquire a lock on the blob by taking a lease and returns the lease ID if successful
func (bl *BlobLocker) AcquireLock(ctx context.Context) (string, error) {

	resp, err := bl.blobLeaseClient.AcquireLease(ctx, int32(bl.lockTTL.Seconds()), nil)
	if err != nil {
		return "", fmt.Errorf("failed to acquire lock for blob %s: %w", bl.blobName, err)
	} else {
		fmt.Println("lease acquired!")
	}

	log.Printf("Lock acquired for blob %s with lease ID: %s", bl.blobName, *resp.LeaseID)
	return *resp.LeaseID, nil
}

// ReleaseLock releases the lock associated with the provided lease ID
func (bl *BlobLocker) ReleaseLock(ctx context.Context) error {
	// Attempt to release the lease
	_, err := bl.blobLeaseClient.ReleaseLease(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to release lock for blob %s with: %w", bl.blobName, err)
	}

	log.Printf("Lock released for blob %s", bl.blobName)
	return nil
}

func handleError(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}
