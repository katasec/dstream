package publishers

import (
	"errors"
	"fmt"
	"log"
	"strings"

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

// // Create returns a ChangePublisher based on the Output.Type in config.
// func (f *ChangePublisherFactory) Create() (ChangePublisher, error) {
// 	switch strings.ToLower(f.config.Output.Type) {
// 	case "eventhub":
// 		log.Println("*** Creating EventHub publisher...")
// 		if f.config.Output.ConnectionString == "" {
// 			return nil, errors.New("EventHub connection string is required")
// 		}
// 		return NewEventHubPublisher(f.config.Output.ConnectionString), nil

// 	case "servicebus":
// 		log.Println("*** Creating ServiceBus publisher...")
// 		if f.config.Output.ConnectionString == "" {
// 			return nil, errors.New("ServiceBus connection string is required")
// 		}
// 		p, err := NewServiceBusPublisher(f.config.Output.ConnectionString, "dstream-instance-1")
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		return p, err

// 	default:
// 		// Default to console if no specific provider is specified
// 		log.Println("Creating Console publisher...")
// 		return NewConsolePublisher(), nil
// 	}
// }

// Create returns a ChangePublisher based on the Output.Type in config.
func (f *ChangePublisherFactory) Create(tableName string) (ChangePublisher, error) {
	switch strings.ToLower(f.config.Output.Type) {
	case "eventhub":
		log.Println("*** Creating EventHub publisher...")
		if f.config.Output.ConnectionString == "" {
			return nil, errors.New("EventHub connection string is required")
		}
		return NewEventHubPublisher(f.config.Output.ConnectionString), nil

	case "servicebus":
		log.Println("*** Creating ServiceBus publisher...")
		if f.config.Output.ConnectionString == "" {
			return nil, errors.New("ServiceBus connection string is required")
		}
		p, err := NewServiceBusPublisher(f.config.Output.ConnectionString, GetServiceBusQueueName(tableName))
		if err != nil {
			log.Fatal(err)
		}
		return p, err

	default:
		// Default to console if no specific provider is specified
		log.Println("Creating Console publisher...")
		return NewConsolePublisher(), nil
	}
}

func GetServiceBusQueueName(tableName string) string {
	tableName = strings.ToLower(tableName)
	return fmt.Sprintf("%s-events", tableName)
}
