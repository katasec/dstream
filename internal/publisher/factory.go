package publisher

import (
	"fmt"
	"strings"

	"github.com/katasec/dstream/internal/publisher/debug/console"
	"github.com/katasec/dstream/internal/publisher/messaging/azure/eventhub"
	"github.com/katasec/dstream/internal/publisher/messaging/azure/servicebus"
	publishertypes "github.com/katasec/dstream/internal/types/publisher"
)

// Factory creates publishers based on configuration
type Factory struct {
	publisherType      publishertypes.Type
	connectionString   string
	dbConnectionString string
}

// NewFactory creates a new publisher factory
func NewFactory(publisherType string, connectionString string, dbConnectionString string) *Factory {
	return &Factory{
		publisherType:      publishertypes.Type(strings.ToLower(publisherType)),
		connectionString:   connectionString,
		dbConnectionString: dbConnectionString,
	}
}

// Create returns a ChangeDataTransport based on the configuration
func (f *Factory) Create(tableName string) (publishertypes.ChangeDataTransport, error) {
	switch f.publisherType {
	// Messaging Publishers
	case publishertypes.AzureServiceBus:
		if f.connectionString == "" {
			return nil, fmt.Errorf("connection string required for Azure Service Bus")
		}
		// Generate the topic name using the consistent naming function
		topicName, err := servicebus.GenTopicName(f.dbConnectionString, tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to generate topic name: %w", err)
		}

		// Create a new ServiceBusChangeDataTransport for the topic
		transport, err := servicebus.NewServiceBusChangeDataTransport(f.connectionString, topicName, false)
		if err != nil {
			return nil, fmt.Errorf("failed to create Azure Service Bus publisher: %w", err)
		}
		return transport, nil

	case publishertypes.AzureEventHub:
		if f.connectionString == "" {
			return nil, fmt.Errorf("connection string required for Azure Event Hub")
		}
		return eventhub.NewEventHubChangeDataTransport(f.connectionString), nil

	// Debug Publishers
	case publishertypes.Console:
		return console.NewConsoleChangeDataTransport(), nil

	// Future implementations
	case publishertypes.AzureBlob:
		return nil, fmt.Errorf("azure blob publisher not yet implemented")
	case publishertypes.AwsS3:
		return nil, fmt.Errorf("aws s3 publisher not yet implemented")
	case publishertypes.SQLDatabase:
		return nil, fmt.Errorf("sql database publisher not yet implemented")
	case publishertypes.MongoDB:
		return nil, fmt.Errorf("mongodb publisher not yet implemented")

	default:
		return nil, fmt.Errorf("unsupported publisher type: %s", f.publisherType)
	}
}
