package messaging

// Publisher defines the interface for publishing messages
type Publisher interface {
	// PublishMessage publishes a message to a topic
	PublishMessage(topic string, message []byte) error

	// EnsureEntityExists ensures that a topic or queue exists, creating it if necessary
	EnsureEntityExists(entityName string) error

	// Close closes the publisher and releases any resources
	Close() error
}
