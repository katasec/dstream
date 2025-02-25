package types

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

// Publisher defines the interface for publishing change data
type Publisher interface {
	// Create creates a new publisher for a specific destination
	Create(destination string) (Publisher, error)

	// PublishMessage publishes a message to a destination
	PublishMessage(ctx context.Context, message interface{}) error

	// EnsureDestinationExists ensures that the destination exists, creating it if necessary
	EnsureDestinationExists(destination string) error

	// Close closes the publisher and releases any resources
	Close() error
}

// ServiceBusPublisher defines the interface for publishing to Azure Service Bus
type ServiceBusPublisher interface {
	Publisher
	// PublishServiceBusMessage publishes a Service Bus message
	PublishServiceBusMessage(ctx context.Context, message *azservicebus.ReceivedMessage) error
}

// Type represents the type of publisher
type Type string

const (
	// Messaging publishers
	AzureServiceBus Type = "azure_service_bus"
	AzureEventHub   Type = "azure_event_hub"

	// Storage publishers
	AzureBlob Type = "azure_blob"
	AwsS3     Type = "aws_s3"

	// Database publishers
	SQLDatabase Type = "sql_database"
	MongoDB     Type = "mongodb"

	// Debug publishers
	Console Type = "console"
	Memory  Type = "memory"
)
