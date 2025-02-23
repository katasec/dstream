// Package messaging provides interfaces for message publishing functionality.
//
// The package defines the core Publisher interface that abstracts message publishing
// capabilities. This allows the system to support multiple messaging backends
// (like Azure Service Bus, Event Hubs, etc.) while maintaining a consistent interface.
//
// Key Components:
//   - Publisher: Core interface for publishing messages to topics
//     * PublishMessage: Publishes a message to a specified topic
//     * EnsureTopicExists: Creates a topic if it doesn't exist
//     * Close: Cleans up resources
package messaging
