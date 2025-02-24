package config

import (
	"github.com/katasec/dstream/internal/publisher"
	"github.com/katasec/dstream/internal/publisher/messaging/azure/servicebus"
	"github.com/katasec/dstream/pkg/cdc"
)

// CreatePublisher creates a publisher for this table configuration
func (t *ResolvedTableConfig) CreatePublisher() (cdc.ChangePublisher, error) {
	// Create a publisher factory
	factory := publisher.NewFactory(
		t.Output.Type,
		t.Output.ConnectionString,
		t.DBConnectionString,
	)

	// Create a publisher for this table
	publisher, err := factory.Create(t.Name)
	if err != nil {
		return nil, err
	}

	// For Service Bus, use the generated topic name
	var destination string
	switch t.Output.Type {
	case string("azure_service_bus"):
		destination = servicebus.GenTopicName(t.DBConnectionString, t.Name)
	default:
		destination = "ingest-queue"
	}

	// Wrap the publisher in an adapter
	return NewPublisherAdapter(publisher, destination), nil
}
