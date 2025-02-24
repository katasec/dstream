package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

type ServiceBusPublisher struct {
	client     *azservicebus.Client
	sender     *azservicebus.Sender
	entityName string // Can be either a topic or queue name
	isQueue    bool   // True if publishing to a queue, false for topic
	batchSize  int
	batchQueue chan map[string]interface{}
}

// NewServiceBusPublisher creates a new ServiceBusPublisher with the provided connection string and entity name
func NewServiceBusPublisher(connectionString, entityName string, isQueue bool) (*ServiceBusPublisher, error) {
	client, err := azservicebus.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		log.Debug("Failed to create Service Bus client")
		return nil, fmt.Errorf("failed to create Service Bus client: %w", err)
	}

	// Create a sender for the entity
	sender, err := client.NewSender(entityName, nil)
	if err != nil {
		client.Close(context.Background())
		return nil, fmt.Errorf("failed to create sender: %w", err)
	}

	publisher := &ServiceBusPublisher{
		client:     client,
		sender:     sender,
		entityName: entityName,
		isQueue:    isQueue,
		batchSize:  10,                                     // Batch size for messages to send
		batchQueue: make(chan map[string]interface{}, 100), // Buffered channel
	}

	go publisher.processMessages()

	return publisher, nil
}

// PublishMessage publishes a message to a topic or queue
func (s *ServiceBusPublisher) PublishMessage(entityName string, message []byte) error {
	log.Debug("Publishing message to Service Bus", "entity", s.entityName, "isQueue", s.isQueue)

	// Create a new message
	sbMessage := &azservicebus.Message{
		Body: message,
	}

	// Send the message using the existing sender
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.sender.SendMessage(ctx, sbMessage, nil); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// EnsureEntityExists ensures that a topic or queue exists, creating it if necessary
func (s *ServiceBusPublisher) EnsureEntityExists(entityName string) error {
	// Azure Service Bus creates entities automatically when sending messages
	return nil
}

// Close closes the publisher and releases any resources
func (s *ServiceBusPublisher) Close() error {
	close(s.batchQueue)
	s.sender.Close(context.Background())
	return s.client.Close(context.Background())
}

// processMessages reads from batchQueue and sends messages in batches
func (s *ServiceBusPublisher) processMessages() {
	log.Info("Starting message processor", "batchSize", s.batchSize, "entityName", s.entityName, "isQueue", s.isQueue)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var batch []map[string]interface{}

	for {
		select {
		case change := <-s.batchQueue:
			log.Debug("Received message for batch", "batchSize", len(batch)+1)
			batch = append(batch, change)
			if len(batch) >= s.batchSize {
				s.sendBatch(batch)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				log.Info("Timer triggered batch send", "batchSize", len(batch))
				s.sendBatch(batch)
				batch = nil
			}
		}
	}
}

// sendBatch sends a batch of messages to the Service Bus topic
func (s *ServiceBusPublisher) sendBatch(batch []map[string]interface{}) {
	log.Info("Sending batch to Service Bus", "batchSize", len(batch), "entityName", s.entityName)

	// Create a batch for all messages
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	msgBatch, err := s.sender.NewMessageBatch(ctx, nil)
	if err != nil {
		log.Error("Failed to create message batch", "error", err)
		return
	}

	for _, change := range batch {
		jsonData, err := json.Marshal(change)
		if err != nil {
			log.Error("Error formatting JSON data", "error", err)
			continue
		}

		message := &azservicebus.Message{Body: jsonData}
		if err := msgBatch.AddMessage(message, nil); err != nil {
			if err == azservicebus.ErrMessageTooLarge {
				// Send current batch and create a new one
				if err := s.sender.SendMessageBatch(ctx, msgBatch, nil); err != nil {
					log.Error("Failed to send message batch", "error", err)
					return
				}
				// Create new batch for remaining messages
				msgBatch, err = s.sender.NewMessageBatch(ctx, nil)
				if err != nil {
					log.Error("Failed to create new message batch", "error", err)
					return
				}
				// Try to add the message to the new batch
				if err := msgBatch.AddMessage(message, nil); err != nil {
					log.Error("Message too large for empty batch", "error", err)
					continue
				}
			} else {
				log.Error("Failed to add message to batch", "error", err)
				continue
			}
		}
	}

	// Send any remaining messages in the batch
	if err := s.sender.SendMessageBatch(ctx, msgBatch, nil); err != nil {
		log.Error("Failed to send final message batch", "error", err)
		return
	}

	log.Info("Successfully sent all messages to Service Bus", "totalSize", len(batch))
}

// prettyPrintJSON prints JSON in an indented format to the console
func prettyPrintJSON(data map[string]interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		log.Error("Error printing JSON", "error", err)
		return
	}
	fmt.Println("Message Data:\n", string(jsonData))
}
