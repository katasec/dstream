package publishers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

type ServiceBusPublisher struct {
	client     *azservicebus.Client
	topicName  string
	batchSize  int
	batchQueue chan map[string]interface{}
}

// NewServiceBusPublisher creates a new ServiceBusPublisher with the provided connection string and topic name
func NewServiceBusPublisher(connectionString, topicName string) (*ServiceBusPublisher, error) {
	client, err := azservicebus.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		log.Debug("Using connection string", "connectionString", connectionString)
		return nil, fmt.Errorf("failed to create Service Bus client: %w", err)
	}

	publisher := &ServiceBusPublisher{
		client:     client,
		topicName:  topicName,
		batchSize:  10,                                     // Batch size for messages to send
		batchQueue: make(chan map[string]interface{}, 100), // Buffered channel
	}

	go publisher.processMessages()

	return publisher, nil
}

// PublishChange sends the change data to the batchQueue for processing
func (s *ServiceBusPublisher) PublishChange(change map[string]interface{}) {
	log.Debug("Queuing message for Service Bus publishing")
	// Print message to console
	prettyPrintJSON(change)

	// Send the message to the channel to be processed in batches
	s.batchQueue <- change
}

// processMessages reads from batchQueue and sends messages in batches
func (s *ServiceBusPublisher) processMessages() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var batch []map[string]interface{}

	for {
		select {
		case change := <-s.batchQueue:
			batch = append(batch, change)
			if len(batch) >= s.batchSize {
				s.sendBatch(batch)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				s.sendBatch(batch)
				batch = nil
			}
		}
	}
}

// sendBatch sends a batch of messages to the Service Bus topic
func (s *ServiceBusPublisher) sendBatch(batch []map[string]interface{}) {
	sender, err := s.client.NewSender(s.topicName, nil)
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
			log.Error("Failed to send message to Service Bus", "error", err)
			continue
		}
		log.Debug("Sent message to Service Bus", "data", string(jsonData))
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
