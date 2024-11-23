package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/katasec/dstream/cdc"
	"github.com/katasec/dstream/config"
	"github.com/katasec/dstream/db"
)

type Server struct {
	config        *config.Config
	lockerFactory *cdc.LockerFactory
	cancelFunc    context.CancelFunc
	wg            *sync.WaitGroup
}

// NewServer initializes the server, loads the configuration, and creates the locker factory
func NewServer() *Server {
	// Load config file
	config, err := config.LoadConfig("dstream.hcl")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	config.CheckConfig()

	// Connect to the database
	dbConn, err := db.Connect(config.DBConnectionString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize LeaseDBManager
	leaseDB := cdc.NewLeaseDBManager(dbConn)

	// Initialize LockerFactory with config and LeaseDBManager
	lockerFactory := cdc.NewLockerFactory(config, leaseDB)

	return &Server{
		config:        config,
		lockerFactory: lockerFactory,
		wg:            &sync.WaitGroup{},
	}
}

// Start initializes the TableMonitoringService and begins monitoring each table in the config
func (s *Server) Start() {
	// Connect to the database
	dbConn, err := db.Connect(s.config.DBConnectionString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	// Create a new context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a new instance of TableMonitoringService
	tableService := cdc.NewTableMonitoringService(dbConn, s.config)

	// Start the TableMonitoringService
	go func() {
		if err := tableService.StartMonitoring(ctx); err != nil {
			log.Printf("Monitoring service error: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	s.handleShutdown(ctx, cancel, tableService)
}

// handleShutdown listens for termination signals and ensures graceful shutdown
func (s *Server) handleShutdown(ctx context.Context, cancel context.CancelFunc, tableService *cdc.TableMonitoringService) {
	// Capture SIGINT and SIGTERM signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	<-signalChan // Wait for a signal
	log.Println("Shutting down gracefully...")

	// Cancel the context to stop all monitoring goroutines
	cancel()

	if tableService != nil {
		// Create a new context for releasing locks
		releaseCtx := context.Background()
		log.Println("Releasing all locks...")
		tableService.ReleaseAllLocks(releaseCtx)
	} else {
		log.Println("No active TableMonitoringService; skipping lock release.")
	}
}
