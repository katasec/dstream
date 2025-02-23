package config

import (
	"encoding/json"

	"github.com/katasec/dstream/pkg/cdc"
	pkg_messaging "github.com/katasec/dstream/pkg/messaging"
)

var _ cdc.ChangePublisher = (*PublisherAdapter)(nil)

// PublisherAdapter adapts a pkg_messaging.Publisher to a cdc.ChangePublisher
type PublisherAdapter struct {
	publisher pkg_messaging.Publisher
	queueName string
}

// NewPublisherAdapter creates a new PublisherAdapter
func NewPublisherAdapter(publisher pkg_messaging.Publisher, queueName string) *PublisherAdapter {
	return &PublisherAdapter{
		publisher: publisher,
		queueName: queueName,
	}
}

// PublishChange implements cdc.ChangePublisher
func (a *PublisherAdapter) PublishChange(data map[string]interface{}) error {
	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Publish using the underlying publisher
	return a.publisher.PublishMessage(a.queueName, jsonData)
}

// Close implements cdc.ChangePublisher
func (a *PublisherAdapter) Close() error {
	return a.publisher.Close()
}
