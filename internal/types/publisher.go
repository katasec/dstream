package types

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

// ChangeDataTransport defines the interface for transporting change data messages
// to various messaging systems like Azure Service Bus, Event Hub, etc.
type ChangeDataTransport interface {
	// Create creates a new transport for a specific destination
	Create(destination string) (ChangeDataTransport, error)

	// PublishBatch publishes a batch of messages to a destination
	// This is the only supported way to publish messages for improved atomicity and resilience
	PublishBatch(ctx context.Context, messages []interface{}) error

	// EnsureDestinationExists ensures that the destination exists, creating it if necessary
	EnsureDestinationExists(destination string) error

	// Close closes the transport and releases any resources
	Close() error
}



// ServiceBusChangeDataTransport defines the interface for transporting change data to Azure Service Bus
type ServiceBusChangeDataTransport interface {
	ChangeDataTransport
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
