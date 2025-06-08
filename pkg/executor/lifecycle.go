package executor

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/katasec/dstream/pkg/logging"
)

// RunWithGracefulShutdown runs a task function with graceful shutdown support.
// It will block until either the task completes or a shutdown signal (SIGINT/SIGTERM) is received.
// If a shutdown signal is received, it will cancel the context and wait for the task to complete cleanup.
func RunWithGracefulShutdown(ctx context.Context, taskFn func(context.Context) error) error {
	log := logging.GetHCLogger()
	
	// Create a cancellable context that can be used for graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)
	
	// Start the task in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- taskFn(ctx)
	}()

	// Wait for either completion or signal
	select {
	case err := <-errCh:
		return err
	case <-sigChan:
		log.Info("Shutdown signal received, stopping gracefully...")
		cancel() // Cancel the context to signal shutdown
		return <-errCh // Wait for the task to finish cleanup
	}
}
