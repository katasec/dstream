package orchestrator

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"time"

	"github.com/katasec/dstream/internal/cdc/locking"
	"github.com/katasec/dstream/internal/cdc/sqlserver"
	"github.com/katasec/dstream/internal/config"
	"github.com/katasec/dstream/internal/logging"
)

var log = logging.GetLogger()

// TableMonitoringOrchestrator coordinates and manages the monitoring of multiple database tables
// for Change Data Capture (CDC) operations. It serves as the central coordinator that:
//  1. Acquires and manages distributed locks to ensure only one instance monitors each table
//  2. Creates and initializes individual table monitors for each configured table
//  3. Coordinates concurrent monitoring across multiple tables using goroutines
//  4. Manages the lifecycle of table monitors, including graceful shutdown
//  5. Handles lock release during shutdown to prevent resource leaks
//
// This orchestrator acts as a higher-level component that delegates the actual CDC monitoring
// to database-specific implementations (e.g., SqlServerTableMonitor).
type TableMonitoringOrchestrator struct {
	db              *sql.DB
	lockerFactory   *locking.LockerFactory
	tablesToMonitor []config.ResolvedTableConfig
	leaseIDs        map[string]string // Map to store lease IDs for each lock
	tableLockers    map[string]locking.DistributedLocker
}

// NewTableMonitoringOrchestrator creates and initializes a new orchestrator instance.
// This constructor:
//  1. Takes a database connection, locker factory, and list of tables to monitor
//  2. Initializes internal tracking maps for lease IDs and table lockers
//  3. Prepares the orchestrator to manage distributed monitoring across tables
//
// The orchestrator doesn't start monitoring until the Start method is explicitly called,
// allowing for proper setup and configuration before monitoring begins.
func NewTableMonitoringOrchestrator(db *sql.DB, lockerFactory *locking.LockerFactory, tablesToMonitor []config.ResolvedTableConfig) *TableMonitoringOrchestrator {
	// Initialize the LeaseDBManager
	//leaseDB := locking.NewLeaseDBManager(db)

	return &TableMonitoringOrchestrator{
		db:              db,
		lockerFactory:   lockerFactory, // Get Locker type from config (for e.g. bloblocker)
		leaseIDs:        make(map[string]string),
		tableLockers:    make(map[string]locking.DistributedLocker),
		tablesToMonitor: tablesToMonitor,
	}
}

// Start initializes and begins the monitoring process for all configured tables.
// This method:
//  1. Creates a distributed lock for each table to ensure exclusive monitoring
//  2. Initializes database-specific monitors (SqlServerTableMonitor) for each table
//  3. Launches concurrent monitoring goroutines for each table
//  4. Manages the lifecycle of all monitors using a WaitGroup
//  5. Staggers the startup of monitors to prevent resource contention
//
// The method blocks until all monitoring goroutines have completed, which typically
// happens only when the context is canceled or an unrecoverable error occurs.
func (t *TableMonitoringOrchestrator) Start(ctx context.Context) error {
	var wg sync.WaitGroup // WaitGroup to ensure goroutines complete

	for _, tableConfig := range t.tablesToMonitor {
		wg.Add(1) // Increment the WaitGroup counter for each table

		// Lock table using a locker from LockerFactory. Add locker to list for releasing locks on exit
		lockName := t.lockerFactory.GetLockName(tableConfig.Name)

		// Create appropriate locker as per config for the table (For e.g. bloblocker)
		tableLocker, err := t.lockerFactory.CreateLocker(lockName)
		if err != nil {
			log.Info("Failed to create locker", "table", tableConfig.Name, "error", err)
			wg.Done()
			continue
		} else {
			// Acquire lock on the table
			_, err := tableLocker.AcquireLock(ctx, lockName)
			if err != nil {
				log.Info("Could not acquire lock on table, exitting", "table", lockName)
				log.Info(err.Error())
				os.Exit(1)
			}
			log.Info("Saving table locker in memory", "table", lockName)

			// Save the locker for later use
			t.tableLockers[lockName] = tableLocker
		}

		// Get polling interval from config
		pollInterval, _ := tableConfig.GetPollInterval()
		maxPollInterval, _ := tableConfig.GetMaxPollInterval()

		// Create a publisher for this table
		publisher, err := tableConfig.CreatePublisher()
		if err != nil {
			log.Info("Failed to create publisher", "table", tableConfig.Name, "error", err)
			continue
		}

		// Create a new SQLServerTableMonitor for this table
		monitor := sqlserver.NewSQLServerTableMonitor(
			t.db,
			tableConfig.Name,
			pollInterval,
			maxPollInterval,
			publisher,
		)

		// Start monitoring this table in a separate goroutine
		go t.monitorTable(&wg, monitor, tableConfig)

		// Stagger the start of each monitor by a short interval
		time.Sleep(500 * time.Millisecond)
	}

	// Wait for all monitoring goroutines to complete
	wg.Wait()
	log.Info("All table monitors have completed.")
	return nil
}

// ReleaseAllLocks safely releases all distributed locks acquired for table monitoring.
// This method is typically called during shutdown to ensure proper cleanup of resources
// and to allow other instances to take over monitoring responsibilities. It iterates
// through all monitored tables and attempts to release their corresponding locks,
// logging the outcome of each release attempt.
func (t *TableMonitoringOrchestrator) ReleaseAllLocks(ctx context.Context) {
	for _, table := range t.tablesToMonitor {
		log.Info("Attempting to release lock", "table", table.Name)
		lockName := t.lockerFactory.GetLockName(table.Name)
		myLocker := t.tableLockers[lockName]

		if myLocker != nil {
			err := myLocker.ReleaseLock(ctx, table.Name, "")
			if err != nil {
				log.Info(err.Error())
			}
		} else {
			log.Info("No lock found", "table", table.Name)
		}
	}
}

// monitorTable manages the lifecycle of a single table monitor within its own goroutine.
// This method:
//  1. Creates a dedicated context for the table monitor
//  2. Initiates the monitoring process for the specific table
//  3. Handles any errors that occur during monitoring
//  4. Ensures the WaitGroup counter is decremented when monitoring completes
//
// It serves as a bridge between the orchestrator and the individual table monitors,
// isolating each monitor in its own execution context.
func (t *TableMonitoringOrchestrator) monitorTable(wg *sync.WaitGroup, monitor *sqlserver.SqlServerTableMonitor, tableConfig config.ResolvedTableConfig) {
	defer wg.Done() // Mark goroutine as done when it completes

	// Create a new context for this table monitor
	ctx := context.Background()

	log.Info("Starting monitor", "table", tableConfig.Name)
	if err := monitor.MonitorTable(ctx); err != nil {
		log.Info("Error monitoring table", "table", tableConfig.Name, "error", err)
	} else {
		log.Info("Monitoring completed", "table", tableConfig.Name)
	}
}
