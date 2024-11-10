package cdc

import (
	"errors"
	"log"

	publishers "github.com/katasec/dstream/cdc/pubishers"
	"github.com/katasec/dstream/config"
)

// ChangePublisherFactory is responsible for creating ChangePublisher instances based on config.
type ChangePublisherFactory struct {
	config *config.Config
}

// NewChangePublisherFactory initializes a new ChangePublisherFactory with the application config.
func NewChangePublisherFactory(config *config.Config) *ChangePublisherFactory {
	return &ChangePublisherFactory{
		config: config,
	}
}

// Create returns a ChangePublisher based on the Output.Type in config.
func (f *ChangePublisherFactory) Create() (publishers.ChangePublisher, error) {
	switch f.config.Output.Type {
	case "EventHub":
		log.Println("Creating EventHub publisher...")
		if f.config.Output.ConnectionString == "" {
			return nil, errors.New("EventHub connection string is required")
		}
		return NewEventHubPublisher(f.config.Output.ConnectionString), nil

	case "ServiceBus":
		log.Println("Creating ServiceBus publisher...")
		if f.config.Output.ConnectionString == "" {
			return nil, errors.New("ServiceBus connection string is required")
		}
		return NewServiceBusPublisher(f.config.Output.ConnectionString), nil

	default:
		// Default to console if no specific provider is specified
		log.Println("Creating Console publisher...")
		return NewConsolePublisher(), nil
	}
}
