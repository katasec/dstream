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
	return &ChangePublisherFactory{
		outputType:         outputType,
		connectionString:   connectionString,
		dbConnectionString: dbConnectionString,
	}
}

// Create returns a Publisher based on the Output.Type in config.
func (f *ChangePublisherFactory) Create(tableName string) (messaging.Publisher, error) {
	switch strings.ToLower(f.outputType) {
	case "eventhub":
		log.Info("Creating EventHub publisher")
		if f.connectionString == "" {
			return nil, errors.New("EventHub connection string is required")
		}
		return NewEventHubPublisher(f.connectionString), nil

	case "servicebus":
		log.Info("Creating ServiceBus publisher")
		if f.connectionString == "" {
			return nil, errors.New("ServiceBus connection string is required")
		}
		// Generate the topic name based on the database and table name
		topicName := GenTopicName(f.dbConnectionString, tableName)
		log.Debug("Using topic", "name", topicName)

		// Create a new ServiceBusPublisher for the topic
		publisher, err := NewServiceBusPublisher(f.connectionString, topicName)
		if err != nil {
			return nil, fmt.Errorf("failed to create ServiceBus publisher for topic %s: %w", topicName, err)
		}
		return publisher, nil

	default:
		// Default to console if no specific provider is specified
		log.Info("Creating Console publisher")
		return NewConsolePublisher(), nil
	}
}
