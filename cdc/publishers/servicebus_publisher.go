package publishers

import (
	"encoding/json"
	"log"
)

// ServiceBusPublisher implements ChangePublisher, sends messages to Azure Service Bus
type ServiceBusPublisher struct {
	connectionString string
}

// NewServiceBusPublisher creates a new ServiceBusPublisher with the provided connection string
func NewServiceBusPublisher(connectionString string) *ServiceBusPublisher {
	return &ServiceBusPublisher{
		connectionString: connectionString,
	}
}

func (s *ServiceBusPublisher) PublishChange(data map[string]interface{}) {
	// Convert data to JSON
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		log.Printf("Error formatting JSON data: %v", err)
		return
	}
	// Placeholder for sending the message to Azure Service Bus
	log.Printf("Sent to Azure Service Bus:\n%s", string(jsonData))
}
