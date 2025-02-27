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

type Ingester struct {
	config        *config.Config
	dbConn        *sql.DB
	lockerFactory *locking.LockerFactory
	// cancelFunc    context.CancelFunc
	wg *sync.WaitGroup
}

// NewIngester initializes the ingester, loads the configuration, and creates the locker factory
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

// Start initializes the TableMonitoringOrchestrator and begins monitoring each table in the config
func (i *Ingester) Start() error {
	defer i.dbConn.Close()

	// Create a new context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Loop until we have tables to monitor
	tablesToMonitor := i.getTablesToMonitor()

	if len(tablesToMonitor) == 0 {
		log.Info("All tables are currently locked, nothing to monitor, exitting.")
		os.Exit(1)
	}

	// Log monitored tables:
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

	// Start Monitoring
	go func() {
		if err := tableMonitoringOrchestrator.Start(ctx); err != nil {
			log.Error("Monitoring orchestrator error", "error", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	i.handleShutdown(cancel, tableMonitoringOrchestrator)
	return nil
}

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
