package publishers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

// ServiceBusPublisher implements ChangePublisher, sends messages to Azure Service Bus
type ServiceBusPublisher struct {
	client           *azservicebus.Client
	queueOrTopicName string
}

// NewServiceBusPublisher creates a new ServiceBusPublisher with the provided connection string
func NewServiceBusPublisher(connectionString, queueOrTopicName string) (*ServiceBusPublisher, error) {
	client, err := azservicebus.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Service Bus client: %w", err)
	}

	return &ServiceBusPublisher{
		client:           client,
		queueOrTopicName: queueOrTopicName,
	}, nil
}

// PublishChange sends the provided data to the Azure Service Bus
func (s *ServiceBusPublisher) PublishChange(data map[string]interface{}) {
	log.Println("ServiceBusPublisher is publishing an event...")
	// Convert data to JSON
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		log.Printf("Error formatting JSON data: %v", err)
		return
	}

	// Create a message with the JSON data
	message := &azservicebus.Message{
		Body: jsonData,
	}

	// Send the message to the Service Bus queue or topic
	sender, err := s.client.NewSender(s.queueOrTopicName, nil)
	if err != nil {
		log.Printf("Failed to create Service Bus sender: %v", err)
		return
	}
	defer sender.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = sender.SendMessage(ctx, message, nil)
	if err != nil {
		log.Printf("Failed to send message to Service Bus: %v", err)
		return
	}

	log.Printf("Sent to Azure Service Bus:\n%s", string(jsonData))
}
