package publisher

// Publisher defines the interface for publishing change data
type Publisher interface {
	// PublishMessage publishes a message to a destination
	PublishMessage(destination string, message []byte) error
	
	// EnsureDestinationExists ensures that the destination exists, creating it if necessary
	EnsureDestinationExists(destination string) error
	
	// Close closes the publisher and releases any resources
	Close() error
}

// Type represents the type of publisher
type Type string

const (
	// Messaging publishers
	AzureServiceBus Type = "azure_service_bus"
	AzureEventHub   Type = "azure_event_hub"
	
	// Storage publishers
	AzureBlob       Type = "azure_blob"
	AwsS3           Type = "aws_s3"
	
	// Database publishers
	SQLDatabase     Type = "sql_database"
	MongoDB         Type = "mongodb"
	
	// Debug publishers
	Console         Type = "console"
	Memory          Type = "memory"
)
