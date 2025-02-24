package service

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

// TableMonitoringService manages monitoring for each table in the config.

type TableMonitoringService struct {
	db              *sql.DB
	lockerFactory   *locking.LockerFactory
	tablesToMonitor []config.ResolvedTableConfig
	leaseIDs        map[string]string // Map to store lease IDs for each lock
	tableLockers    map[string]locking.DistributedLocker
}

func NewTableMonitoringService(db *sql.DB, lockerFactory *locking.LockerFactory, tablesToMonitor []config.ResolvedTableConfig) *TableMonitoringService {
	// Initialize the LeaseDBManager
	//leaseDB := locking.NewLeaseDBManager(db)

	return &TableMonitoringService{
		db:              db,
		lockerFactory:   lockerFactory, // Get Locker type from config (for e.g. bloblocker)
		leaseIDs:        make(map[string]string),
		tableLockers:    make(map[string]locking.DistributedLocker),
		tablesToMonitor: tablesToMonitor,
	}
}

// Start initializes and starts monitoring for each table in the config.
func (t *TableMonitoringService) Start(ctx context.Context) error {
	var wg sync.WaitGroup // WaitGroup to ensure goroutines complete

	for _, tableConfig := range t.tablesToMonitor {
		wg.Add(1) // Increment the WaitGroup counter for each table

		// Lock table using a locker from LockerFactory. Add locker to list for releasing locks on exit
		lockName := tableConfig.Name + ".lock"
		tableLocker, err := t.lockerFactory.CreateLocker(lockName)
		if err != nil {
			log.Info("Failed to create locker", "table", tableConfig.Name, "error", err)
			wg.Done()
			continue
		} else {
			_, err := tableLocker.AcquireLock(ctx, lockName)
			if err != nil {
				log.Info("Could not acquire lock on table, exitting", "table", lockName)
				log.Info(err.Error())
				os.Exit(1)
			}
			log.Info("Saving table locker in memory", "table", lockName)
			t.tableLockers[lockName] = tableLocker
		}

		// Get polling interval from config
		pollInterval, _ := tableConfig.GetPollInterval()
		maxPollInterval, _ := tableConfig.GetMaxPollInterval()

		// Initialize SQLServerMonitor for each table with poll intervals and the correct publisher.
		// Create a publisher for this table
		publisher, err := tableConfig.CreatePublisher()
		if err != nil {
			log.Info("Failed to create publisher", "table", tableConfig.Name, "error", err)
			continue
		}

		monitor := sqlserver.NewSQLServerTableMonitor(
			t.db,
			tableConfig.Name,
			pollInterval,
			maxPollInterval,
			publisher,
		)

		// Start monitoring each table as a separate goroutine using the helper function
		go t.monitorTable(&wg, monitor, tableConfig)

		// Stagger the start of each monitor by a short interval
		time.Sleep(500 * time.Millisecond)
	}

	// Wait for all monitoring goroutines to complete
	wg.Wait()
	log.Info("All table monitors have completed.")
	return nil
}

func (t *TableMonitoringService) ReleaseAllLocks(ctx context.Context) {
	for _, table := range t.tablesToMonitor {
		log.Info("Attempting to release lock", "table", table.Name)
		lockName := table.Name + ".lock"
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

func (t *TableMonitoringService) monitorTable(wg *sync.WaitGroup, monitor *sqlserver.SqlServerTableMonitor, tableConfig config.ResolvedTableConfig) {
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
