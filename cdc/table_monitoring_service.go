package cdc

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/katasec/dstream/cdc/lockers"
	"github.com/katasec/dstream/cdc/publishers"
	"github.com/katasec/dstream/config"
)

// TableMonitoringService manages monitoring for each table in the config.

type TableMonitoringService struct {
	db              *sql.DB
	config          *config.Config
	lockerFactory   *lockers.LockerFactory
	tablesToMonitor []config.TableConfig
	leaseIDs        map[string]string // Map to store lease IDs for each lock
	// renewalCancelFuncs map[string]context.CancelFunc // Map to store cancel functions for renewal goroutines
	// mu           sync.Mutex // Mutex to synchronize access to maps
	tableLockers map[string]lockers.DistributedLocker
}

// NewTableMonitoringService initializes a new TableMonitoringService.
// func NewTableMonitoringService(db *sql.DB, config *config.Config) *TableMonitoringService {
// 	return &TableMonitoringService{
// 		db:                 db,
// 		config:             config,
// 		lockerFactory:      NewLockerFactory(config),
// 		leaseIDs:           make(map[string]string),
// 		renewalCancelFuncs: make(map[string]context.CancelFunc),
// 	}
// }

func NewTableMonitoringService(db *sql.DB, config *config.Config, tablesToMonitor []config.TableConfig) *TableMonitoringService {
	// Initialize the LeaseDBManager
	//leaseDB := lockers.NewLeaseDBManager(db)

	return &TableMonitoringService{
		db:              db,
		config:          config,
		lockerFactory:   lockers.NewLockerFactory(config), // Get Locker type from config (for e.g. bloblocker)
		leaseIDs:        make(map[string]string),
		tableLockers:    make(map[string]lockers.DistributedLocker),
		tablesToMonitor: tablesToMonitor,
	}
}

// StartMonitoring initializes and starts monitoring for each table in the config.
func (t *TableMonitoringService) StartMonitoring(ctx context.Context) error {
	var wg sync.WaitGroup // WaitGroup to ensure goroutines complete

	// Initialize ChangePublisherFactory
	publisherFactory := publishers.NewChangePublisherFactory(t.config)

	//for _, tableConfig := range t.config.Tables {
	for _, tableConfig := range t.tablesToMonitor {
		wg.Add(1) // Increment the WaitGroup counter for each table

		// Create publisher per table (Servicebus or EventHub etc)
		publisher, err := publisherFactory.Create(tableConfig.Name)
		if err != nil {
			log.Printf("Error creating publisher for table %s: %v", tableConfig.Name, err)
			wg.Done()
			continue
		}

		// Lock table using a locker from LockerFactory. Add locker to list for releasing locks on exit
		lockName := tableConfig.Name + ".lock"
		tableLocker, err := t.lockerFactory.CreateLocker(lockName)
		if err != nil {
			log.Printf("Failed to create locker for table %s: %v", tableConfig.Name, err)
			wg.Done()
			continue
		} else {
			tableLocker.AcquireLock(ctx, lockName)
			log.Println("Saving table locker for:", lockName)
			t.tableLockers[lockName] = tableLocker
		}

		// Get polling interval from config
		pollInterval, _ := tableConfig.GetPollInterval()
		maxPollInterval, _ := tableConfig.GetMaxPollInterval()

		// Initialize SQLServerMonitor for each table with poll intervals and the correct publisher.
		monitor := NewSQLServerMonitor(
			t.db,
			tableConfig.Name,
			pollInterval,
			maxPollInterval,
			publisher,
		)

		// Start monitoring each table as a separate goroutine using the helper function
		go t.monitorTable(ctx, &wg, monitor, tableConfig, lockName, tableLocker)

		// Stagger the start of each monitor by a short interval
		time.Sleep(500 * time.Millisecond)
	}

	// Wait for all monitoring goroutines to complete
	wg.Wait()
	log.Println("All table monitors have completed.")
	return nil
}

func (t *TableMonitoringService) ReleaseAllLocks(ctx context.Context) {
	for _, table := range t.config.Tables {
		log.Printf("Attempting to release lock for:%s \n", table.Name)
		lockName := table.Name + ".lock"
		myLocker := t.tableLockers[lockName]

		if myLocker != nil {
			err := myLocker.ReleaseLock(ctx, table.Name, "")
			if err != nil {
				log.Println(err.Error())
			}
		} else {
			log.Printf("No lock found for %s\n", table.Name)
		}
	}
}

func (t *TableMonitoringService) monitorTable(ctx context.Context, wg *sync.WaitGroup, monitor *SQLServerMonitor, tableConfig config.TableConfig, lockName string, locker lockers.DistributedLocker) {
	defer wg.Done() // Mark goroutine as done when it completes

	log.Printf("Starting monitor for table: %s", tableConfig.Name)
	if err := monitor.MonitorTable(); err != nil {
		log.Printf("Error monitoring table %s: %v", tableConfig.Name, err)
	} else {
		log.Printf("Monitoring completed for table %s", tableConfig.Name)
	}

	// Start lock renewal
	locker.StartLockRenewal(ctx, lockName)
}
