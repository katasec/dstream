package publishers

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

// ServiceBusPublisher is an asynchronous publisher for sending messages to Azure Service Bus
type ServiceBusPublisher struct {
	client       *azservicebus.Client
	tableName    string
	messages     chan map[string]interface{}
	batchSize    int
	batchTimeout time.Duration
}

// NewServiceBusPublisher creates a new ServiceBusPublisher with the provided connection string and topic/queue name
func NewServiceBusPublisher(connectionString, tableName string) (*ServiceBusPublisher, error) {
	client, err := azservicebus.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, err
	}

	// Initialize the publisher with a buffered channel and batch settings
	publisher := &ServiceBusPublisher{
		client:       client,
		tableName:    tableName,
		messages:     make(chan map[string]interface{}, 100), // Buffer size of 100; adjust as needed
		batchSize:    10,                                     // Max number of messages per batch
		batchTimeout: 10 * time.Second,                       // Max wait time for batching
	}

	// Start the background goroutine to process messages
	go publisher.processMessages()

	return publisher, nil
}

// PublishChange queues a message to be sent to Azure Service Bus
func (s *ServiceBusPublisher) PublishChange(data map[string]interface{}) {
	select {
	case s.messages <- data:
		// Successfully queued message
		log.Println("Queued message for asynchronous publishing.")
	default:
		// Channel is full; log or handle overflow
		log.Println("Warning: message queue is full; dropping message")
	}
}

// processMessages runs as a background goroutine to send messages in batches
func (s *ServiceBusPublisher) processMessages() {
	batch := make([]*azservicebus.Message, 0, s.batchSize)
	timer := time.NewTimer(s.batchTimeout)
	defer timer.Stop()

	for {
		select {
		case data := <-s.messages:
			// Convert data to JSON
			jsonData, err := json.MarshalIndent(data, "", "    ")
			if err != nil {
				log.Printf("Error formatting JSON data: %v", err)
				continue
			}

			// Log the message to console as it’s being added to the batch
			log.Printf("Console Output - Message:\n%s", string(jsonData))

			// Add message to batch
			message := &azservicebus.Message{Body: jsonData}
			batch = append(batch, message)

			// Send batch if it reaches the batch size limit
			if len(batch) >= s.batchSize {
				s.sendBatch(batch)
				batch = batch[:0] // Reset batch
				timer.Reset(s.batchTimeout)
			}

		case <-timer.C:
			// Send any remaining messages in the batch if timeout expires
			if len(batch) > 0 {
				s.sendBatch(batch)
				batch = batch[:0] // Reset batch
			}
			// Reset the timer
			timer.Reset(s.batchTimeout)
		}
	}
}

// sendBatch sends a batch of messages to the Service Bus
func (s *ServiceBusPublisher) sendBatch(batch []*azservicebus.Message) {
	sender, err := s.client.NewSender(s.tableName, nil)
	if err != nil {
		log.Printf("Failed to create Service Bus sender: %v", err)
		return
	}
	defer sender.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Send each message in the batch
	for _, message := range batch {
		if err := sender.SendMessage(ctx, message, nil); err != nil {
			log.Printf("Failed to send message to Service Bus: %v", err)
		} else {
			log.Println("Sent message to Service Bus.")
		}
	}
}
