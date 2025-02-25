package router

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/katasec/dstream/internal/config"
	"github.com/katasec/dstream/internal/logging"
	"github.com/katasec/dstream/internal/publisher"
	"github.com/katasec/dstream/internal/publisher/messaging/azure/servicebus"
)

var log = logging.GetLogger()

// Router handles routing messages from the ingest queue to their destination topics
type Router struct {
	config     *config.Config
	publisher  publisher.Publisher
	publishers map[string]publisher.Publisher // Cache of topic publishers
	lock       sync.RWMutex                  // Lock for publishers map
}

// NewRouter creates a new Router instance
func NewRouter(cfg *config.Config) (*Router, error) {
	// Initialize the publisher based on configuration
	factory := publisher.NewFactory(
		cfg.Publisher.Output.Type,
		cfg.Publisher.Output.ConnectionString,
		cfg.Ingester.DBConnectionString,
	)

	basePublisher, err := factory.Create("")
	if err != nil {
		return nil, fmt.Errorf("failed to create base publisher: %w", err)
	}

	r := &Router{
		config:     cfg,
		publisher:  basePublisher,
		publishers: make(map[string]publisher.Publisher),
	}

	// Pre-create publishers for all tables in config
	for _, table := range cfg.Ingester.Tables {
		// Generate the topic name
		topicName, err := servicebus.GenTopicName(cfg.Ingester.DBConnectionString, table.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to generate topic name for table %s: %w", table.Name, err)
		}

		// Ensure topic exists
		if err := basePublisher.EnsureDestinationExists(topicName); err != nil {
			return nil, fmt.Errorf("failed to ensure topic exists for table %s: %w", table, err)
		}

		// Create publisher for this topic
		topicPublisher, err := basePublisher.Create(topicName)
		if err != nil {
			return nil, fmt.Errorf("failed to create publisher for table %s: %w", table, err)
		}

		r.publishers[topicName] = topicPublisher
		log.Debug("Created publisher for topic", "topic", topicName)
	}

	return r, nil
}

// Start begins routing messages from the ingest queue to their destinations
func (r *Router) Start() error {
	ctx, cancel := context.WithCancel(context.Background())

	// Start processing in a goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := r.routeMessages(ctx); err != nil {
			log.Error("Error processing changes", "error", err)
		}
	}()

	// Wait for shutdown signal
	r.handleShutdown(cancel)
	wg.Wait()

	return nil
}

// routeMessages continuously routes messages from the ingest queue to their destinations
func (r *Router) routeMessages(ctx context.Context) error {
	// Create a Service Bus receiver for the ingest queue
	receiver, err := servicebus.NewReceiver(r.config.Ingester.Queue.ConnectionString, r.config.Ingester.Queue.Name)
	if err != nil {
		return fmt.Errorf("failed to create queue receiver: %w", err)
	}
	defer receiver.Close(ctx)



	log.Info("Started routing messages from ingest queue")

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Receive message from the queue
			msg, err := receiver.ReceiveMessages(ctx, 1, nil)
			if err != nil {
				log.Error("Failed to receive message", "error", err)
				continue
			}

			// No messages available
			if len(msg) == 0 {
				continue
			}

			// Parse the message body
			message := msg[0]
			var messageData struct {
				Metadata struct {
					Destination string `json:"Destination"`
				} `json:"metadata"`
			}

			if err := json.Unmarshal(message.Body, &messageData); err != nil {
				log.Error("Failed to parse message body", "error", err)
				continue
			}

			if messageData.Metadata.Destination == "" {
				log.Error("Message missing destination in metadata")
				continue
			}

			destinationTopic := messageData.Metadata.Destination

			// Ensure the topic exists
			if err := r.publisher.EnsureDestinationExists(destinationTopic); err != nil {
				log.Error("Failed to ensure topic exists", "topic", destinationTopic, "error", err)
				continue
			}

			// Get the cached publisher for this topic
			r.lock.RLock()
			topicPublisher, exists := r.publishers[destinationTopic]
			r.lock.RUnlock()

			if !exists {
				log.Error("No publisher found for topic", "topic", destinationTopic)
				continue
			}

			// Forward the message to the destination topic
			if err := topicPublisher.PublishMessage(ctx, msg[0]); err != nil {
				log.Error("Failed to publish message to topic", "topic", destinationTopic, "error", err)
				continue
			}

			// Complete the message (remove from queue)
			if err := receiver.CompleteMessage(ctx, msg[0]); err != nil {
				log.Error("Failed to complete message", "error", err)
				continue
			}

			log.Info("Successfully routed message", "destination", destinationTopic)
		}
	}
}

// handleShutdown waits for interrupt signal and initiates graceful shutdown
func (r *Router) handleShutdown(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Info("Received shutdown signal, initiating graceful shutdown")
	cancel()
}
