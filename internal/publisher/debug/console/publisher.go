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

// ConsoleChangeDataTransport implements publisher.ChangeDataTransport, outputs change data messages to console
type ConsoleChangeDataTransport struct{}

// NewConsoleChangeDataTransport creates a new ConsoleChangeDataTransport instance
func NewConsoleChangeDataTransport() *ConsoleChangeDataTransport {
	return &ConsoleChangeDataTransport{}
}

// Create creates a new transport for a specific destination
func (p *ConsoleChangeDataTransport) Create(destination string) (types.ChangeDataTransport, error) {
	return NewConsoleChangeDataTransport(), nil
}

// PublishBatch publishes a batch of messages to the console
func (p *ConsoleChangeDataTransport) PublishBatch(ctx context.Context, messages []interface{}) error {
	if len(messages) == 0 {
		return nil // Nothing to publish
	}
	
	log.Info("Publishing batch to console", "batchSize", len(messages))
	
	// Process each message in the batch
	for i, message := range messages {
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

		log.Info(fmt.Sprintf("Message %d/%d:", i+1, len(messages)))
		log.Info(prettyJSON.String())
	}
	
	return nil
}

// EnsureDestinationExists is a no-op for console transport
func (p *ConsoleChangeDataTransport) EnsureDestinationExists(destination string) error {
	return nil
}

// Close is a no-op for console transport
func (p *ConsoleChangeDataTransport) Close() error {
	return nil
}
