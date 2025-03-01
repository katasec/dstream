package ingester

import (
	"context"
	"database/sql"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/katasec/dstream/internal/logging"

	"github.com/katasec/dstream/internal/cdc/locking"
	"github.com/katasec/dstream/internal/cdc/orchestrator"
	"github.com/katasec/dstream/internal/config"
	"github.com/katasec/dstream/internal/db"
)

var log = logging.GetLogger()

// Ingester is the top-level component in the dstream architecture that:
//  1. Initializes the system configuration and database connections
//  2. Manages the lifecycle of the TableMonitoringOrchestrator
//  3. Handles system signals for graceful shutdown
//  4. Coordinates the overall data ingestion process
//
// The Ingester serves as the entry point and controller for the entire
// data streaming pipeline, delegating the actual table monitoring work
// to the TableMonitoringOrchestrator while maintaining responsibility
// for system-wide concerns like configuration and shutdown procedures.
type Ingester struct {
	config        *config.Config
	dbConn        *sql.DB
	lockerFactory *locking.LockerFactory
	// cancelFunc    context.CancelFunc
	wg *sync.WaitGroup
}

// NewIngester initializes the ingester, loads the configuration, and creates the locker factory.
// This constructor:
//  1. Loads and validates the system configuration
//  2. Establishes the database connection
//  3. Initializes the distributed locking factory
//  4. Prepares the Ingester to coordinate the data ingestion process
//
// The Ingester is created in a ready state but doesn't start any processing
// until the Start method is explicitly called.
func NewIngester() *Ingester {

	config := config.NewConfig()
	config.CheckConfig()

	// Connect to the database
	dbConn, err := db.Connect(config.Ingester.DBConnectionString)
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Initialize LeaseDBManager
	//leaseDB := locking.NewLeaseDBManager(dbConn)

	// Initialize LockerFactory with config and LeaseDBManager
	configType := config.Ingester.Locks.Type
	connectionString := config.Ingester.Locks.ConnectionString
	containerName := config.Ingester.Locks.ContainerName
	lockerFactory := locking.NewLockerFactory(configType, connectionString, containerName)

	return &Ingester{
		config:        config,
		dbConn:        dbConn,
		lockerFactory: lockerFactory,
		wg:            &sync.WaitGroup{},
	}
}

// Start initializes the TableMonitoringOrchestrator and begins monitoring each table in the config.
// This method:
//  1. Creates a cancellable context to control the monitoring lifecycle
//  2. Identifies available tables that can be monitored (not locked by other instances)
//  3. Initializes the TableMonitoringOrchestrator with the tables to monitor
//  4. Launches the orchestrator in a separate goroutine to prevent blocking
//  5. Sets up signal handling for graceful shutdown
//
// The goroutine used to start the orchestrator is critical to the design as it:
//  - Allows the main thread to proceed to signal handling
//  - Enables non-blocking orchestration of the monitoring process
//  - Maintains the ability to propagate cancellation signals to all monitoring activities
//
// The method returns only after a shutdown signal is received and processed.
func (i *Ingester) Start() error {
	defer i.dbConn.Close()

	// Create a new context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Get list of tables to monitor (unlocked tables only)
	tablesToMonitor := i.getTablesToMonitor()
	if len(tablesToMonitor) == 0 {
		log.Info("All tables are currently locked, nothing to monitor, exitting.")
		os.Exit(1)
	}

	// Log the tables to be monitored
	log.Info("The following tables will be monitored")
	for _, table := range tablesToMonitor {
		log.Info("Monitoring table", "name", table.Name)
	}

	// Create table monitoring orchestrator for those tables
	locksConfig := i.config.Ingester.Locks
	lockerFactory := locking.NewLockerFactory(
		locksConfig.Type,
		locksConfig.ConnectionString,
		locksConfig.ContainerName,
	)
	tableMonitoringOrchestrator := orchestrator.NewTableMonitoringOrchestrator(i.dbConn, lockerFactory, tablesToMonitor)

	// Start Monitoring in a separate goroutine to avoid blocking the main thread
	// This allows the program to continue to the handleShutdown function
	// while monitoring happens in the background
	go func() {
		if err := tableMonitoringOrchestrator.Start(ctx); err != nil {
			log.Error("Monitoring orchestrator error", "error", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	i.handleShutdown(cancel, tableMonitoringOrchestrator)
	return nil
}

// getTablesToMonitor identifies which tables from the configuration can be monitored.
// This method:
//  1. Retrieves the list of configured tables
//  2. Checks which tables are already locked by other instances
//  3. Filters out locked tables to prevent duplicate monitoring
//  4. Returns only the tables that are available for monitoring
//
// This function plays an important role in the distributed architecture by:
//  - Ensuring tables are not monitored by multiple instances simultaneously
//  - Allowing the system to scale horizontally across multiple instances
//  - Providing automatic work distribution among available instances
func (i *Ingester) getTablesToMonitor() []config.ResolvedTableConfig {

	// Get list of table names from config
	var tablesToMonitor []config.ResolvedTableConfig
	tableNames := make([]string, len(i.config.Ingester.Tables))
	for i, table := range i.config.Ingester.Tables {
		tableNames[i] = table.Name
	}

	// Create the locker defined in the config file (For e.g. blob locker)
	configType := i.config.Ingester.Locks.Type
	connectionString := i.config.Ingester.Locks.ConnectionString
	containerName := i.config.Ingester.Locks.ContainerName
	lockerFactory := locking.NewLockerFactory(configType, connectionString, containerName)

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
	for _, table := range i.config.Ingester.Tables {
		lockName := lockerFactory.GetLockName(table.Name)
		if lockedTableMap[lockName] {
			log.Info("Table is locked and being monitored by another process", "table", table.Name)
			continue
		}

		tablesToMonitor = append(tablesToMonitor, table)
	}

	// Return the list of table to monitor
	return tablesToMonitor
}

// handleShutdown listens for termination signals and ensures graceful shutdown.
// This method:
//  1. Sets up signal handling for SIGINT and SIGTERM
//  2. Blocks until a termination signal is received
//  3. Cancels the monitoring context to stop all monitoring goroutines
//  4. Ensures all distributed locks are properly released
//
// This function is a critical part of the Ingester's architecture as it:
//  - Provides a clean shutdown mechanism for the entire system
//  - Ensures resources are properly released (particularly distributed locks)
//  - Prevents resource leaks and allows other instances to take over monitoring
//  - Completes the lifecycle management responsibility of the Ingester
func (i *Ingester) handleShutdown(cancel context.CancelFunc, tableService *orchestrator.TableMonitoringOrchestrator) {
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
		log.Info("No active TableMonitoringOrchestrator; skipping lock release.")
	}
}
