package config

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/katasec/dstream/internal/publisher"
	"github.com/katasec/dstream/internal/publisher/messaging/azure/servicebus"
	"github.com/katasec/dstream/pkg/cdc"
)

var _ cdc.ChangePublisher = (*PublisherAdapter)(nil)

// PublisherAdapter adapts a pkg_messaging.Publisher to a cdc.ChangePublisher
type PublisherAdapter struct {
	publisher    publisher.Publisher
	queueName    string
	dbConnString string
}

// NewPublisherAdapter creates a new PublisherAdapter
func NewPublisherAdapter(publisher publisher.Publisher, queueName string, dbConnString string) *PublisherAdapter {
	return &PublisherAdapter{
		publisher:    publisher,
		queueName:    queueName,
		dbConnString: dbConnString,
	}
}

// PublishChange implements cdc.ChangePublisher
func (a *PublisherAdapter) PublishChange(data map[string]interface{}) (<-chan bool, error) {
	// Create a channel to signal successful publish
	doneChan := make(chan bool, 1)
	// Get database connection string and table name from metadata
	metadata, ok := data["metadata"].(map[string]interface{})
	if !ok {
		close(doneChan)
		return doneChan, fmt.Errorf("metadata not found in change data")
	}

	// Generate the final destination topic name
	tableName, ok := metadata["TableName"].(string)
	if !ok {
		close(doneChan)
		return doneChan, fmt.Errorf("TableName not found in metadata")
	}

	// Add both the immediate destination (ingest-queue) and final destination topic
	metadata["IngestQueue"] = a.queueName

	// Generate the destination topic name
	destination, err := servicebus.GenTopicName(a.dbConnString, tableName)
	if err != nil {
		close(doneChan)
		return doneChan, fmt.Errorf("failed to generate topic name: %w", err)
	}
	metadata["Destination"] = destination

	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		close(doneChan)
		return doneChan, err
	}

	// Publish using the underlying publisher
	if err := a.publisher.PublishMessage(context.Background(), jsonData); err != nil {
		close(doneChan)
		return doneChan, err
	}

	// Signal successful publish
	doneChan <- true
	close(doneChan)

	return doneChan, nil
}

// Close implements cdc.ChangePublisher
func (a *PublisherAdapter) Close() error {
	return a.publisher.Close()
}
