package config

import (
	"github.com/katasec/dstream/internal/messaging"
	"github.com/katasec/dstream/pkg/cdc"
)

// CreatePublisher creates a publisher for this table configuration
func (t *ResolvedTableConfig) CreatePublisher() (cdc.ChangePublisher, error) {
	// Create a publisher factory
	factory := messaging.NewChangePublisherFactory(
		t.Output.Type,
		t.Output.ConnectionString,
		t.DBConnectionString,
	)

	// Create a publisher for this table
	publisher, err := factory.Create(t.Name)
	if err != nil {
		return nil, err
	}

	// Wrap the publisher in an adapter
	return NewPublisherAdapter(publisher, "ingest-queue"), nil
}
