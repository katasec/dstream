// Package messaging provides implementations of message publishing interfaces.
//
// The package contains concrete implementations of the messaging.Publisher interface
// for different messaging systems:
//   - Azure Service Bus
//   - Azure Event Hubs
//   - Console (for debugging)
//
// It also provides factory methods for creating publishers based on configuration.
//
// Key Components:
//   - ServiceBusPublisher: Azure Service Bus implementation
//   - EventHubPublisher: Azure Event Hubs implementation
//   - ConsolePublisher: Simple console-based implementation for testing
//   - ChangePublisherFactory: Factory for creating publishers based on config
package messaging
