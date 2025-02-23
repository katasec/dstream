package messaging

import (
	"errors"
	"fmt"
	"strings"

	"github.com/katasec/dstream/pkg/messaging"
)

// ChangePublisherFactory is responsible for creating ChangePublisher instances based on config.
type ChangePublisherFactory struct {
	outputType         string
	connectionString   string
	dbConnectionString string
}

func NewChangePublisherFactory(outputType string, connectionString string, dbConnectionString string) *ChangePublisherFactory {
	log.Info("Creating publisher factory", "outputType", outputType, "connectionString", connectionString)
	return &ChangePublisherFactory{
		outputType:         outputType,
		connectionString:   connectionString,
		dbConnectionString: dbConnectionString,
	}
}

// Create returns a Publisher based on the Output.Type in config.
func (f *ChangePublisherFactory) Create(tableName string) (messaging.Publisher, error) {
	log.Info("Creating publisher", "outputType", f.outputType, "connectionString", f.connectionString)
	switch strings.ToLower(f.outputType) {
	case "eventhub":
		log.Info("Creating EventHub publisher")
		if f.connectionString == "" {
			return nil, errors.New("EventHub connection string is required")
		}
		return NewEventHubPublisher(f.connectionString), nil

	case "azure_service_bus":
		log.Info("Creating ServiceBus publisher")
		if f.connectionString == "" {
			return nil, errors.New("ServiceBus connection string is required")
		}
		// Create a new ServiceBusPublisher for the ingest queue
		queueName := "ingest-queue"
		log.Info("Creating ServiceBus queue publisher", "name", queueName, "connectionString", f.connectionString)

		// Create a new ServiceBusPublisher for the queue, explicitly setting isQueue to true
		publisher, err := NewServiceBusPublisher(f.connectionString, queueName, true)
		if err != nil {
			return nil, fmt.Errorf("failed to create ServiceBus publisher for queue %s: %w", queueName, err)
		}
		return publisher, nil

	default:
		// Default to console if no specific provider is specified
		log.Info("Creating Console publisher")
		return NewConsolePublisher(), nil
	}
}
