package publishers

import (
	"encoding/json"
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
		log.Error("Error formatting JSON data", "error", err)
		return
	}
	// Placeholder for sending the message to Azure Event Hub
	log.Info("Sent to Azure Event Hub", "data", string(jsonData))
}
