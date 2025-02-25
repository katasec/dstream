package console

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/katasec/dstream/internal/logging"
	"github.com/katasec/dstream/internal/types"
)

var log = logging.GetLogger()

// Publisher implements publisher.Publisher, outputs to console
type Publisher struct{}

// NewPublisher creates a new ConsolePublisher instance
func NewPublisher() *Publisher {
	return &Publisher{}
}

// Create creates a new publisher for a specific destination
func (p *Publisher) Create(destination string) (types.Publisher, error) {
	return NewPublisher(), nil
}

// PublishMessage publishes a message to the console
func (p *Publisher) PublishMessage(ctx context.Context, message interface{}) error {
	var data []byte
	switch msg := message.(type) {
	case []byte:
		data = msg
	default:
		return fmt.Errorf("unsupported message type: %T", message)
	}

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		log.Error("Failed to format JSON", "error", err)
		return err
	}

	log.Info("Publishing message to console")
	log.Info(prettyJSON.String())
	return nil
}

// EnsureDestinationExists is a no-op for console publisher
func (p *Publisher) EnsureDestinationExists(destination string) error {
	return nil
}

// Close is a no-op for console publisher
func (p *Publisher) Close() error {
	return nil
}
