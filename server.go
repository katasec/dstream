package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/katasec/dstream/cdc"
	"github.com/katasec/dstream/cdc/lockers"
	"github.com/katasec/dstream/config"
	"github.com/katasec/dstream/db"
)

type Server struct {
	config        *config.Config
	dbConn        *sql.DB
	lockerFactory *lockers.LockerFactory
	// cancelFunc    context.CancelFunc
	wg *sync.WaitGroup
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
	//leaseDB := lockers.NewLeaseDBManager(dbConn)

	// Initialize LockerFactory with config and LeaseDBManager
	lockerFactory := lockers.NewLockerFactory(config)

	return &Server{
		config:        config,
		dbConn:        dbConn,
		lockerFactory: lockerFactory,
		wg:            &sync.WaitGroup{},
	}
}

// Start initializes the TableMonitoringService and begins monitoring each table in the config
func (s *Server) Start() {
	defer s.dbConn.Close()

	// Create a new context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Loop until we have tables to monitor
	tablesToMonitor := s.getTablestoMonitor()

	if len(tablesToMonitor) == 0 {
		log.Println("All tables are currently locked, nothing to monitor, exitting.")
		os.Exit(1)
	}

	// Log monitored tables:
	log.Println("The following tables will be monitored:")
	for _, table := range tablesToMonitor {
		log.Println("Table Name:", table.Name)
	}

	// Create table monitoring service for those tables
	tableService := cdc.NewTableMonitoringService(s.dbConn, s.config, tablesToMonitor)

	// Start Monitoring
	go func() {
		if err := tableService.StartMonitoring(ctx); err != nil {
			log.Printf("Monitoring service error: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	s.handleShutdown(cancel, tableService)
}

func (s *Server) getTablestoMonitor() []config.TableConfig {
	tablesToMonitor := []config.TableConfig{}

	// Create the locker defined in the config HCL
	lockerFactory := lockers.NewLockerFactory(s.config)

	// From the tables in config, find tables that are locked
	// and already being monitored by another process
	lockedTables, _ := lockerFactory.GetLockedTables()

	// Convert lockedTables to a set for efficient lookup
	lockedTableSet := make(map[string]struct{})
	for _, locked := range lockedTables {
		lockedTableSet[locked] = struct{}{}
	}

	// Iterate the tables in config
	for _, table := range s.config.Tables {
		lockName := table.Name + ".lock"

		// Check if lockName is in lockedTableSet
		if _, exists := lockedTableSet[lockName]; exists {
			// Skip if lockName is already in lockedTableSet
			continue
		}

		// Add to tablesToMonitor if not locked
		tablesToMonitor = append(tablesToMonitor, table)
	}

	return tablesToMonitor
}

// handleShutdown listens for termination signals and ensures graceful shutdown
func (s *Server) handleShutdown(cancel context.CancelFunc, tableService *cdc.TableMonitoringService) {
	// Capture SIGINT and SIGTERM signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	<-signalChan // Wait for a signal
	log.Println("Ctrl-C detected, shutting down gracefully...")

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
