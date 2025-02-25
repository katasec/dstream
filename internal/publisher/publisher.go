package publisher

// Re-export types from internal/types package
import (
	"github.com/katasec/dstream/internal/types"
)

type Publisher = types.Publisher
type ServiceBusPublisher = types.ServiceBusPublisher
type Type = types.Type

const (
	// Messaging publishers
	AzureServiceBus = types.AzureServiceBus
	AzureEventHub   = types.AzureEventHub

	// Storage publishers
	AzureBlob = types.AzureBlob
	AwsS3     = types.AwsS3

	// Database publishers
	SQLDatabase = types.SQLDatabase
	MongoDB     = types.MongoDB

	// Debug publishers
	Console = types.Console
	Memory  = types.Memory
)
