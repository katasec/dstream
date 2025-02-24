package console

import (
	"bytes"
	"encoding/json"

	"github.com/katasec/dstream/internal/logging"
)

var log = logging.GetLogger()

// Publisher implements publisher.Publisher, outputs to console
type Publisher struct{}

// NewPublisher creates a new ConsolePublisher instance
func NewPublisher() *Publisher {
	return &Publisher{}
}

// PublishMessage publishes a message to the console
func (p *Publisher) PublishMessage(destination string, message []byte) error {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, message, "", "  "); err != nil {
		log.Error("Failed to format JSON", "error", err)
		return err
	}

	log.Info("Publishing message to console", "destination", destination)
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
