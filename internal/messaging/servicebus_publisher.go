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
	entityName string // Can be either a topic or queue name
	isQueue    bool   // True if publishing to a queue, false for topic
	batchSize  int
	batchQueue chan map[string]interface{}
}

// NewServiceBusPublisher creates a new ServiceBusPublisher with the provided connection string and entity name
func NewServiceBusPublisher(connectionString, entityName string, isQueue bool) (*ServiceBusPublisher, error) {
	client, err := azservicebus.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		log.Debug("Using connection string", "connectionString", connectionString)
		return nil, fmt.Errorf("failed to create Service Bus client: %w", err)
	}

	publisher := &ServiceBusPublisher{
		client:     client,
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
	sender, err := s.client.NewSender(s.entityName, nil)
	if err != nil {
		return fmt.Errorf("failed to create sender: %w", err)
	}
	defer sender.Close(context.TODO())

	// Create a new message
	sbMessage := &azservicebus.Message{
		Body: message,
	}

	// Send the message
	if err := sender.SendMessage(context.TODO(), sbMessage, nil); err != nil {
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
	return s.client.Close(context.TODO())
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
	sender, err := s.client.NewSender(s.entityName, nil)
	if err != nil {
		log.Error("Failed to create Service Bus sender", "error", err)
		return
	}
	defer sender.Close(context.Background())

	for _, change := range batch {
		jsonData, err := json.Marshal(change)
		if err != nil {
			log.Error("Error formatting JSON data", "error", err)
			continue
		}

		message := &azservicebus.Message{Body: jsonData}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := sender.SendMessage(ctx, message, nil); err != nil {
			log.Error("Failed to send message to Service Bus", "error", err, "data", string(jsonData))
			continue
		}
		log.Info("Sent message to Service Bus", "size", len(jsonData))
	}
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
