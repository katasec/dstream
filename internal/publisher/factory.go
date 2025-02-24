package publisher

import (
	"fmt"
	"strings"

	"github.com/katasec/dstream/internal/publisher/debug/console"
	"github.com/katasec/dstream/internal/publisher/messaging/azure/eventhub"
	"github.com/katasec/dstream/internal/publisher/messaging/azure/servicebus"
)

// Factory creates publishers based on configuration
type Factory struct {
	publisherType     Type
	connectionString  string
	dbConnectionString string
}

// NewFactory creates a new publisher factory
func NewFactory(publisherType string, connectionString string, dbConnectionString string) *Factory {
	return &Factory{
		publisherType:     Type(strings.ToLower(publisherType)),
		connectionString:  connectionString,
		dbConnectionString: dbConnectionString,
	}
}

// Create returns a Publisher based on the configuration
func (f *Factory) Create(tableName string) (Publisher, error) {
	switch f.publisherType {
	// Messaging Publishers
	case AzureServiceBus:
		if f.connectionString == "" {
			return nil, fmt.Errorf("connection string required for Azure Service Bus")
		}
		// Generate the topic name using the consistent naming function
		topicName := servicebus.GenTopicName(f.dbConnectionString, tableName)
		
		// Create a new ServiceBusPublisher for the topic
		publisher, err := servicebus.NewPublisher(f.connectionString, topicName, false)
		if err != nil {
			return nil, fmt.Errorf("failed to create Azure Service Bus publisher: %w", err)
		}
		return publisher, nil

	case AzureEventHub:
		if f.connectionString == "" {
			return nil, fmt.Errorf("connection string required for Azure Event Hub")
		}
		return eventhub.NewPublisher(f.connectionString), nil

	// Debug Publishers
	case Console:
		return console.NewPublisher(), nil

	// Future implementations
	case AzureBlob:
		return nil, fmt.Errorf("azure blob publisher not yet implemented")
	case AwsS3:
		return nil, fmt.Errorf("aws s3 publisher not yet implemented")
	case SQLDatabase:
		return nil, fmt.Errorf("sql database publisher not yet implemented")
	case MongoDB:
		return nil, fmt.Errorf("mongodb publisher not yet implemented")
	
	default:
		return nil, fmt.Errorf("unsupported publisher type: %s", f.publisherType)
	}
}
