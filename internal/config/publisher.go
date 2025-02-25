package config

import (
	"github.com/katasec/dstream/internal/publisher"
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

	// Always use ingest-queue as immediate destination
	destination := "ingest-queue"

	// Wrap the publisher in an adapter with both queue name and connection string
	return NewPublisherAdapter(publisher, destination, t.DBConnectionString), nil
}
