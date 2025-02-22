package cdc

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"time"

	"github.com/katasec/dstream/cdc/lockers"
	"github.com/katasec/dstream/config"
)

// TableMonitoringService manages monitoring for each table in the config.

type TableMonitoringService struct {
	db              *sql.DB
	lockerFactory   *lockers.LockerFactory
	tablesToMonitor []config.ResolvedTableConfig
	leaseIDs        map[string]string // Map to store lease IDs for each lock
	tableLockers    map[string]lockers.DistributedLocker
}

func NewTableMonitoringService(db *sql.DB, lockerFactory *lockers.LockerFactory, tablesToMonitor []config.ResolvedTableConfig) *TableMonitoringService {
	// Initialize the LeaseDBManager
	//leaseDB := lockers.NewLeaseDBManager(db)

	return &TableMonitoringService{
		db:              db,
		lockerFactory:   lockerFactory, // Get Locker type from config (for e.g. bloblocker)
		leaseIDs:        make(map[string]string),
		tableLockers:    make(map[string]lockers.DistributedLocker),
		tablesToMonitor: tablesToMonitor,
	}
}

// StartMonitoring initializes and starts monitoring for each table in the config.
func (t *TableMonitoringService) StartMonitoring(ctx context.Context) error {
	var wg sync.WaitGroup // WaitGroup to ensure goroutines complete

	for _, tableConfig := range t.tablesToMonitor {
		wg.Add(1) // Increment the WaitGroup counter for each table

		// Lock table using a locker from LockerFactory. Add locker to list for releasing locks on exit
		lockName := tableConfig.Name + ".lock"
		tableLocker, err := t.lockerFactory.CreateLocker(lockName)
		if err != nil {
			log.Info("Failed to create locker for table %s: %v", tableConfig.Name, err)
			wg.Done()
			continue
		} else {
			_, err := tableLocker.AcquireLock(ctx, lockName)
			if err != nil {
				log.Info("Could not acquire lock on table: %s, exitting.", lockName)
				log.Info(err.Error())
				os.Exit(1)
			}
			log.Info("Saving table locker in memory for:", lockName)
			t.tableLockers[lockName] = tableLocker
		}

		// Get polling interval from config
		pollInterval, _ := tableConfig.GetPollInterval()
		maxPollInterval, _ := tableConfig.GetMaxPollInterval()

		// Initialize SQLServerMonitor for each table with poll intervals and the correct publisher.
		monitor := NewSQLServerTableMonitor(
			t.db,
			tableConfig.Name,
			pollInterval,
			maxPollInterval,
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
		log.Info("Attempting to release lock for:%s \n", table.Name)
		lockName := table.Name + ".lock"
		myLocker := t.tableLockers[lockName]

		if myLocker != nil {
			err := myLocker.ReleaseLock(ctx, table.Name, "")
			if err != nil {
				log.Info(err.Error())
			}
		} else {
			log.Info("No lock found for %s\n", table.Name)
		}
	}
}

func (t *TableMonitoringService) monitorTable(wg *sync.WaitGroup, monitor *SqlServerTableMonitor, tableConfig config.ResolvedTableConfig) {
	defer wg.Done() // Mark goroutine as done when it completes

	// Create a new context for this table monitor
	ctx := context.Background()

	log.Info("Starting monitor for table: %s", tableConfig.Name)
	if err := monitor.MonitorTable(ctx); err != nil {
		log.Info("Error monitoring table %s: %v", tableConfig.Name, err)
	} else {
		log.Info("Monitoring completed for table %s", tableConfig.Name)
	}

}
