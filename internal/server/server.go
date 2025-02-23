package server

import (
	"context"
	"database/sql"
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
	// config, err := config.LoadConfig("dstream.hcl")
	// if err != nil {
	// 	log.Fatalf("Error loading config: %v", err)
	// }

	config := config.NewConfig()

	config.CheckConfig()

	// Connect to the database
	dbConn, err := db.Connect(config.Ingester.DBConnectionString)
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Initialize LeaseDBManager
	//leaseDB := lockers.NewLeaseDBManager(dbConn)

	// Initialize LockerFactory with config and LeaseDBManager
	configType := config.Ingester.Locks.Type
	connectionString := config.Ingester.Locks.ConnectionString
	containerName := config.Ingester.Locks.ContainerName
	lockerFactory := lockers.NewLockerFactory(configType, connectionString, containerName)

	return &Server{
		config:        config,
		dbConn:        dbConn,
		lockerFactory: lockerFactory,
		wg:            &sync.WaitGroup{},
	}
}

// Start initializes the TableMonitoringService and begins monitoring each table in the config
func (s *Server) Start() error {
	defer s.dbConn.Close()

	// Create a new context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Loop until we have tables to monitor
	tablesToMonitor := s.getTablesToMonitor()

	if len(tablesToMonitor) == 0 {
		log.Info("All tables are currently locked, nothing to monitor, exitting.")
		os.Exit(1)
	}

	// Log monitored tables:
	log.Info("The following tables will be monitored")
	for _, table := range tablesToMonitor {
		log.Info("Monitoring table", "name", table.Name)
	}

	// Create table monitoring service for those tables
	locksConfig := s.config.Ingester.Locks
	lockerFactory := lockers.NewLockerFactory(
		locksConfig.Type,
		locksConfig.ConnectionString,
		locksConfig.ContainerName,
	)
	tableService := cdc.NewTableMonitoringService(s.dbConn, lockerFactory, tablesToMonitor)

	// Start Monitoring
	go func() {
		if err := tableService.StartMonitoring(ctx); err != nil {
			log.Error("Monitoring service error", "error", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	s.handleShutdown(cancel, tableService)
	return nil
}

func (s *Server) getTablesToMonitor() []config.ResolvedTableConfig {

	// Get list of table names from config
	var tablesToMonitor []config.ResolvedTableConfig
	tableNames := make([]string, len(s.config.Ingester.Tables))
	for i, table := range s.config.Ingester.Tables {
		tableNames[i] = table.Name
	}

	// Create the locker defined in the config file (For e.g. blob locker)
	configType := s.config.Ingester.Locks.Type
	connectionString := s.config.Ingester.Locks.ConnectionString
	containerName := s.config.Ingester.Locks.ContainerName
	lockerFactory := lockers.NewLockerFactory(configType, connectionString, containerName)

	// Pass tableNames to locker factory to see if they are locked
	lockedTables, err := lockerFactory.GetLockedTables(tableNames)
	if err != nil {
		log.Error("Error checking locked tables", "error", err)
		return tablesToMonitor
	}

	// Convert locked tables to a map for efficient lookup
	lockedTableMap := make(map[string]bool)
	for _, lockedTable := range lockedTables {
		lockedTableMap[lockedTable] = true
	}

	// Filter out the locked tables from the list
	for _, table := range s.config.Ingester.Tables {
		lockName := table.Name + ".lock"
		if lockedTableMap[lockName] {
			log.Info("Table is locked and being monitored by another process", "table", table.Name)
			continue
		}

		tablesToMonitor = append(tablesToMonitor, table)
	}

	// Return the list of table to monitor
	return tablesToMonitor
}

// handleShutdown listens for termination signals and ensures graceful shutdown
func (s *Server) handleShutdown(cancel context.CancelFunc, tableService *cdc.TableMonitoringService) {
	// Capture SIGINT and SIGTERM signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signalChan) // Unregister the signal handlers
	defer close(signalChan)       // Clean up the channel

	<-signalChan // Wait for a signal
	log.Info("Ctrl-C detected, shutting down gracefully...")

	// Cancel the context to stop all monitoring goroutines
	cancel()

	if tableService != nil {
		// Create a new context for releasing locks
		releaseCtx := context.Background()
		log.Info("Releasing all locks...")
		tableService.ReleaseAllLocks(releaseCtx)
	} else {
		log.Info("No active TableMonitoringService; skipping lock release.")
	}
}
