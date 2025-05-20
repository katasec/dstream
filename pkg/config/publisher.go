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

	// Create a transport for this table
	transport, err := factory.Create(t.Name)
	if err != nil {
		return nil, err
	}

	// Always use ingest-queue as immediate destination
	destination := "ingest-queue"

	// Create a ChangeDataPublisher with the transport
	return NewChangeDataPublisher(transport, destination, t.DBConnectionString), nil
}
