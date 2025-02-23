package messaging

// Publisher defines the interface for publishing messages
type Publisher interface {
	// PublishMessage publishes a message to a topic
	PublishMessage(topic string, message []byte) error

	// EnsureTopicExists ensures that a topic exists, creating it if necessary
	EnsureTopicExists(topic string) error

	// Close closes the publisher and releases any resources
	Close() error
}
