package processor

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/katasec/dstream/internal/config"
	"github.com/katasec/dstream/internal/logging"
	"github.com/katasec/dstream/internal/publisher"

)

var log = logging.GetLogger()

// Processor handles consuming changes from a queue and publishing them to destinations
type Processor struct {
	config    *config.Config
	publisher publisher.Publisher
}

// NewProcessor creates a new Processor instance
func NewProcessor(cfg *config.Config) *Processor {
	return &Processor{
		config: cfg,
	}
}

// Start begins processing changes from the queue
func (p *Processor) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize the publisher based on configuration
	factory := publisher.NewFactory(
		p.config.Publisher.Output.Type,
		p.config.Publisher.Output.ConnectionString,
		p.config.Ingester.DBConnectionString,
	)

	var err error
	p.publisher, err = factory.Create("")
	if err != nil {
		return err
	}

	// Start processing in a goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := p.processChanges(ctx); err != nil {
			log.Error("Error processing changes", "error", err)
		}
	}()

	// Wait for shutdown signal
	p.handleShutdown(cancel)
	wg.Wait()

	return nil
}

// processChanges continuously processes changes from the queue
func (p *Processor) processChanges(ctx context.Context) error {
	// TODO: Implement queue consumer and change processing logic
	<-ctx.Done()
	return nil
}

// handleShutdown waits for interrupt signal and initiates graceful shutdown
func (p *Processor) handleShutdown(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Info("Received shutdown signal, initiating graceful shutdown")
	cancel()
}
