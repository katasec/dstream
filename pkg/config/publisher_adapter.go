package config

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/katasec/dstream/internal/publisher/messaging/azure/servicebus"
	publishertypes "github.com/katasec/dstream/internal/types/publisher"
	"github.com/katasec/dstream/pkg/cdc"
)

var _ cdc.ChangePublisher = (*ChangeDataPublisher)(nil)

// ChangeDataPublisher handles the publishing of CDC change data to messaging systems
// It enriches CDC data with routing information and uses a ChangeDataTransport for delivery
type ChangeDataPublisher struct {
	transport    publishertypes.ChangeDataTransport
	queueName    string
	dbConnString string
}

// NewChangeDataPublisher creates a new ChangeDataPublisher
func NewChangeDataPublisher(transport publishertypes.ChangeDataTransport, queueName string, dbConnString string) *ChangeDataPublisher {
	return &ChangeDataPublisher{
		transport:    transport,
		queueName:    queueName,
		dbConnString: dbConnString,
	}
}

// PublishChanges implements cdc.ChangePublisher
func (p *ChangeDataPublisher) PublishChanges(changes []map[string]interface{}) (<-chan bool, error) {
	// Create a channel to signal successful publish
	doneChan := make(chan bool, 1)

	// If no changes, return immediately with success
	if len(changes) == 0 {
		doneChan <- true
		close(doneChan)
		return doneChan, nil
	}

	// Process all changes in the batch
	var messages []interface{}

	// First, validate and enrich all messages
	for _, data := range changes {
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
		metadata["IngestQueue"] = p.queueName

		// Generate the destination topic name
		destination, err := servicebus.GenTopicName(p.dbConnString, tableName)
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

		// Add the JSON data to our batch
		messages = append(messages, jsonData)
	}

	// Publish the entire batch using the underlying transport's batch API
	if err := p.transport.PublishBatch(context.Background(), messages); err != nil {
		close(doneChan)
		return doneChan, err
	}

	// Signal successful publish
	doneChan <- true
	close(doneChan)

	return doneChan, nil
}

// Close implements cdc.ChangePublisher
func (p *ChangeDataPublisher) Close() error {
	return p.transport.Close()
}
