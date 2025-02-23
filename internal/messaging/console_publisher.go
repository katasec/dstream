package messaging

import (
	"bytes"
	"encoding/json"
)

// ConsolePublisher implements ChangePublisher, outputs to console
type ConsolePublisher struct{}

// NewConsolePublisher creates a new ConsolePublisher instance
func NewConsolePublisher() *ConsolePublisher {
	return &ConsolePublisher{}
}

// PublishMessage publishes a message to a topic
func (c *ConsolePublisher) PublishMessage(topic string, message []byte) error {
	// Try to pretty-print if it's JSON
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, message, "", "    "); err == nil {
		log.Info("Console Publisher output", "topic", topic, "data", prettyJSON.String())
	} else {
		log.Info("Console Publisher output", "topic", topic, "data", string(message))
	}
	return nil
}

// EnsureEntityExists ensures that a topic or queue exists, creating it if necessary
func (c *ConsolePublisher) EnsureEntityExists(entityName string) error {
	// Console publisher doesn't need entities
	return nil
}

// Close closes the publisher and releases any resources
func (c *ConsolePublisher) Close() error {
	// Nothing to close for console publisher
	return nil
}
