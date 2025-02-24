package servicebus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/katasec/dstream/internal/logging"
)

var log = logging.GetLogger()

// Publisher implements publisher.Publisher, sends messages to Azure Service Bus
type Publisher struct {
	client     *azservicebus.Client
	sender     *azservicebus.Sender
	entityName string
	isQueue    bool
	batchSize  int
	batchQueue chan map[string]interface{}
}

// NewPublisher creates a new ServiceBusPublisher with the provided connection string and entity name
func NewPublisher(connectionString, entityName string, isQueue bool) (*Publisher, error) {
	client, err := azservicebus.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		log.Debug("Failed to create Service Bus client")
		return nil, fmt.Errorf("failed to create Service Bus client: %w", err)
	}

	sender, err := client.NewSender(entityName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create sender: %w", err)
	}

	p := &Publisher{
		client:     client,
		sender:     sender,
		entityName: entityName,
		isQueue:    isQueue,
		batchSize:  100,
		batchQueue: make(chan map[string]interface{}, 1000),
	}

	// Start processing messages in batches
	go p.processMessages()

	return p, nil
}

// PublishMessage publishes a message to a topic or queue
func (p *Publisher) PublishMessage(entityName string, message []byte) error {
	// Create a new Service Bus message
	sbMessage := &azservicebus.Message{
		Body: message,
	}

	// Send the message
	if err := p.sender.SendMessage(context.Background(), sbMessage, nil); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// EnsureDestinationExists ensures that a topic or queue exists, creating it if necessary
func (p *Publisher) EnsureDestinationExists(entityName string) error {
	// Implementation moved to utility functions
	return nil
}

// Close closes the publisher and releases any resources
func (p *Publisher) Close() error {
	if p.sender != nil {
		return p.sender.Close(context.Background())
	}
	return nil
}

// processMessages reads from batchQueue and sends messages in batches
func (p *Publisher) processMessages() {
	var batch []map[string]interface{}
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case msg := <-p.batchQueue:
			batch = append(batch, msg)
			if len(batch) >= p.batchSize {
				p.sendBatch(batch)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				p.sendBatch(batch)
				batch = nil
			}
		}
	}
}

// sendBatch sends a batch of messages to the Service Bus topic
func (p *Publisher) sendBatch(batch []map[string]interface{}) {
	ctx := context.Background()

	// Convert all messages to Service Bus messages first
	var messages []*azservicebus.Message
	for _, msg := range batch {
		data, err := json.Marshal(msg)
		if err != nil {
			log.Error("Failed to marshal message", "error", err)
			continue
		}

		messages = append(messages, &azservicebus.Message{
			Body: data,
		})
	}

	// Send messages in batches
	for i := 0; i < len(messages); i += p.batchSize {
		end := i + p.batchSize
		if end > len(messages) {
			end = len(messages)
		}

		// Send this batch
		batchMessages := messages[i:end]
		if err := p.sender.SendMessage(ctx, batchMessages[0], nil); err != nil {
			log.Error("Failed to send message", "error", err)
		}
	}
}
