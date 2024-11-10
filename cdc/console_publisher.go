package cdc

import (
	"encoding/json"
	"log"
)

// ConsolePublisher implements ChangePublisher, outputs to console
type ConsolePublisher struct{}

// NewConsolePublisher creates a new ConsolePublisher instance
func NewConsolePublisher() *ConsolePublisher {
	return &ConsolePublisher{}
}

func (c *ConsolePublisher) PublishChange(data map[string]interface{}) {
	// Replace the operation code with a human-readable action
	if opCode, ok := data["Operation"].(int); ok {
		data["Operation"] = TranslateOperation(opCode)
	}

	// Pretty-print the JSON data
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		log.Printf("Error formatting JSON data: %v", err)
		return
	}

	log.Printf("Console Publisher:\n%s", string(jsonData))
}
