package cdc

import (
	"encoding/json"
	"log"
)

// EventHubPublisher implements ChangePublisher, sends messages to Azure Event Hub
type EventHubPublisher struct {
	connectionString string
}

// NewEventHubPublisher creates a new EventHubPublisher with the provided connection string
func NewEventHubPublisher(connectionString string) *EventHubPublisher {
	return &EventHubPublisher{
		connectionString: connectionString,
	}
}

func (e *EventHubPublisher) PublishChange(data map[string]interface{}) {
	// Convert data to JSON
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		log.Printf("Error formatting JSON data: %v", err)
		return
	}
	// Placeholder for sending the message to Azure Event Hub
	log.Printf("Sent to Azure Event Hub:\n%s", string(jsonData))
}
