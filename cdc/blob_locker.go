package cdc

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/lease"
)

type BlobLocker struct {
	client          *azblob.Client
	container       string
	lockTTL         time.Duration
	blockblobClient *blockblob.Client
	blobLeaseClient *lease.BlobClient
	leaseID         string
}
