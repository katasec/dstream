package lockers

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/lease"
	"github.com/katasec/dstream/config"
)

type BlobLocker struct {
	containerName string
	lockTTL       time.Duration
	lockName      string

	azblobClient    *azblob.Client
	blobLeaseClient *lease.BlobClient
	ctx             context.Context
}

func NewBlobLocker(connectionString, containerName, lockName string, lockTTL time.Duration) (*BlobLocker, error) {

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

	lockTTL = -1 * time.Second
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
	log.Printf("Attempting to acquire lock for blob %s", bl.lockName)

	// Acquire Lease
	resp, err := bl.blobLeaseClient.AcquireLease(bl.ctx, int32(bl.lockTTL.Seconds()), nil)
	if err != nil {
		if strings.Contains(err.Error(), "There is already a lease present") {
			log.Printf("Table %s is already locked. Skipping...", bl.lockName)
			return "", nil
		}
		return "", fmt.Errorf("failed to acquire lock for blob %s: %w", bl.lockName, err)
	}

	log.Printf("Lock acquired for blob %s with Lease ID: %s", bl.lockName, *resp.LeaseID)

	return "", nil
}

func (bl *BlobLocker) RenewLock(ctx context.Context, lockName string) error {
	_, err := bl.blobLeaseClient.RenewLease(bl.ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to renew lock for blob %s: %w", lockName, err)
	}

	log.Printf("Lock renewed for blob %s", lockName)
	return nil
}

// ReleaseLock releases the lock associated with the provided lease ID for the specified blob (lockName)
func (bl *BlobLocker) ReleaseLock(tx context.Context, lockName string, leaseID string) error {

	_, err := bl.blobLeaseClient.ReleaseLease(bl.ctx, &lease.BlobReleaseOptions{})
	if err != nil {
		return fmt.Errorf("failed to release lock for blob %s: %w", bl.lockName, err)
	} else {
		log.Printf("Lock released successfully for blob %s !\n", bl.lockName)
	}
	return nil
}

func (bl *BlobLocker) StartLockRenewal(ctx context.Context, lockName string) {
	log.Printf("Starting lock renewal for blob %s", lockName)
	go func() {
		ticker := time.NewTicker(bl.lockTTL / 2)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := bl.RenewLock(bl.ctx, bl.lockName); err != nil {
					log.Printf("Failed to renew lock for blob %s: %v", lockName, err)
				}
			case <-ctx.Done():
				log.Printf("Stopping lock renewal for blob %s", lockName)
				return
			}
		}
	}()
}

// GetLockedTables Iterates through all the blobs in the container to find locked tables
func (bl *BlobLocker) GetLockedTables() []string {
	lockedTables := []string{}
	config := config.NewConfig()

	// Create blob pager
	containerName := config.Locks.ContainerName
	containerClient := bl.azblobClient.ServiceClient().NewContainerClient(containerName)
	pager := containerClient.NewListBlobsFlatPager(nil)

	// User pager to iterate through blob store pages
	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			log.Fatalf("Failed to list blobs: %v", err)
		}
		// for each page, iterate through blob items
		for _, blob := range page.Segment.BlobItems {
			blobName := *blob.Name
			log.Printf("Blob: %s\n", blobName)

			// Get the BlobClient for the current blob
			blobClient := containerClient.NewBlobClient(blobName)

			// Fetch the blob's properties
			resp, err := blobClient.GetProperties(context.TODO(), nil)
			if err != nil {
				log.Printf("Failed to get properties for blob %s: %v\n", blobName, err)
				continue
			}

			// Check the lease status and state
			leaseStatus := resp.LeaseStatus
			leaseState := resp.LeaseState

			log.Printf("  Lease Status: %s\n", *leaseStatus)
			log.Printf("  Lease State: %s\n", *leaseState)

			if *leaseStatus == "locked" && *leaseState == "leased" {
				log.Printf("  -> The blob is leased.\n")
				lockedTables = append(lockedTables, blobName)
			} else {
				log.Printf("  -> The blob is not leased.\n")
			}
		}
	}

	return lockedTables
}

func GetBlobLockerLockedTables(config *config.Config) []string {
	lockedTables := []string{}
	// Create azblobClient and create container
	azblobClient, err := azblob.NewClientFromConnectionString(config.Locks.ConnectionString, nil)
	if err != nil {
		log.Println("failed to create Azure Blob client: %w, exitting.", err)
		os.Exit(1)
	}

	// Create blob pager
	containerClient := azblobClient.ServiceClient().NewContainerClient(config.Locks.ContainerName)
	pager := containerClient.NewListBlobsFlatPager(nil)

	// User pager to iterate through blob store pages
	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			log.Fatalf("Failed to list blobs: %v", err)
		}
		// for each page, iterate through blob items
		for _, blob := range page.Segment.BlobItems {
			blobName := *blob.Name
			log.Printf("Blob: %s\n", blobName)

			// Get the BlobClient for the current blob
			blobClient := containerClient.NewBlobClient(blobName)

			// Fetch the blob's properties
			resp, err := blobClient.GetProperties(context.TODO(), nil)
			if err != nil {
				log.Printf("Failed to get properties for blob %s: %v\n", blobName, err)
				continue
			}

			// Check the lease status and state
			leaseStatus := resp.LeaseStatus
			leaseState := resp.LeaseState

			fmt.Printf("  Lease Status: %s\n", *leaseStatus)
			fmt.Printf("  Lease State: %s\n", *leaseState)

			if *leaseStatus == "locked" && *leaseState == "leased" {
				fmt.Printf("  -> The blob is leased.\n")
				lockedTables = append(lockedTables, blobName)
			} else {
				fmt.Printf("  -> The blob is not leased.\n")
			}
		}
	}

	return lockedTables
}
