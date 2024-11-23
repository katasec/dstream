package cdc

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/katasec/dstream/cdc/publishers"
	"github.com/katasec/dstream/config"
)

// TableMonitoringService manages monitoring for each table in the config.

type TableMonitoringService struct {
	db                 *sql.DB
	config             *config.Config
	lockerFactory      *LockerFactory
	leaseIDs           map[string]string             // Map to store lease IDs for each lock
	renewalCancelFuncs map[string]context.CancelFunc // Map to store cancel functions for renewal goroutines
	mu                 sync.Mutex                    // Mutex to synchronize access to maps
	leaseDB            *LeaseDBManager
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

func NewTableMonitoringService(db *sql.DB, config *config.Config) *TableMonitoringService {
	// Initialize the LeaseDBManager
	leaseDB := NewLeaseDBManager(db)

	return &TableMonitoringService{
		db:            db,
		config:        config,
		lockerFactory: NewLockerFactory(config, leaseDB), // Pass leaseDB to NewLockerFactory
		leaseDB:       leaseDB,                           // Assign leaseDB to the TableMonitoringService
		leaseIDs:      make(map[string]string),
	}
}

// StartMonitoring initializes and starts monitoring for each table in the config.
// StartMonitoring initializes and starts monitoring for each table in the config.
func (t *TableMonitoringService) StartMonitoring(ctx context.Context) error {
	var wg sync.WaitGroup // WaitGroup to ensure goroutines complete

	// Initialize ChangePublisherFactory
	publisherFactory := publishers.NewChangePublisherFactory(t.config)

	// Get tables that are not locked
	unlockedTables := t.GetUnlockedTables()

	for _, tableConfig := range unlockedTables {
		wg.Add(1) // Increment the WaitGroup counter for each table

		pollInterval, _ := tableConfig.GetPollInterval()
		maxPollInterval, _ := tableConfig.GetMaxPollInterval()

		// Create the appropriate publisher for the table
		publisher, err := publisherFactory.Create(tableConfig.Name)
		if err != nil {
			log.Printf("Error creating publisher for table %s: %v", tableConfig.Name, err)
			wg.Done()
			continue
		}

		// Define the lock name for the table
		lockName := tableConfig.Name + ".lock"

		// Check if a lease ID already exists in the database
		existingLeaseID, err := t.leaseDB.GetLeaseID(lockName)
		if err != nil {
			log.Printf("Failed to retrieve lease ID for table %s: %v", tableConfig.Name, err)
			wg.Done()
			continue
		}
		if existingLeaseID != "" {
			log.Printf("Skipping table %s as it is already locked (Lease ID: %s).", tableConfig.Name, existingLeaseID)
			wg.Done()
			continue
		}

		// Create a locker for the table using the LockerFactory
		locker, err := t.lockerFactory.CreateLocker(lockName)
		if err != nil {
			log.Printf("Failed to create locker for table %s: %v", tableConfig.Name, err)
			wg.Done()
			continue
		}

		// Acquire the lock
		leaseID, err := locker.AcquireLock(ctx, lockName)
		if err != nil {
			log.Printf("Failed to acquire lock for table %s: %v", tableConfig.Name, err)
			wg.Done()
			continue
		}
		if leaseID == "" {
			log.Printf("Skipping table %s as it is already locked.", tableConfig.Name)
			wg.Done()
			continue
		}

		// Store the lease ID in memory and persist it in the database
		t.mu.Lock()
		t.leaseIDs[lockName] = leaseID
		t.mu.Unlock()

		if err := t.leaseDB.StoreLeaseID(lockName, leaseID); err != nil {
			log.Printf("Failed to persist lease ID for table %s: %v", tableConfig.Name, err)
			wg.Done()
			continue
		}

		// Initialize SQLServerMonitor for each table with poll intervals and the correct publisher.
		monitor := NewSQLServerMonitor(
			t.db,
			tableConfig.Name,
			pollInterval,
			maxPollInterval,
			publisher,
		)

		// Start monitoring each table as a separate goroutine
		go func(monitor *SQLServerMonitor, tableConfig config.TableConfig, locker DistributedLocker) {
			defer wg.Done() // Mark goroutine as done when it completes

			// Start lock renewal
			locker.StartLockRenewal(ctx, lockName)

			log.Printf("Starting monitor for table: %s", tableConfig.Name)
			if err := monitor.MonitorTable(); err != nil {
				log.Printf("Error monitoring table %s: %v", tableConfig.Name, err)
			} else {
				log.Printf("Monitoring completed for table %s", tableConfig.Name)
			}
		}(monitor, tableConfig, locker)

		// Stagger the start of each monitor by a short interval
		time.Sleep(500 * time.Millisecond)
	}

	// Wait for all monitoring goroutines to complete
	wg.Wait()
	log.Println("All table monitors have completed.")
	return nil
}

// GetUnlockedTables retrieves tables that are not locked
func (t *TableMonitoringService) GetUnlockedTables() []config.TableConfig {
	var unlockedTables []config.TableConfig
	ctx := context.TODO()

	for _, tableConfig := range t.config.Tables {
		lockName := tableConfig.Name + ".lock"

		// Use LockerFactory to create a locker for each table
		locker, err := t.lockerFactory.CreateLocker(lockName)
		if err != nil {
			log.Printf("Failed to create locker for table %s: %v", tableConfig.Name, err)
			continue
		}

		// Try to acquire a lock temporarily to check if the table is locked
		leaseID, err := locker.AcquireLock(ctx, lockName)
		if err != nil {
			fmt.Printf("Could not acquire lock for %s: %s\n", tableConfig.Name, err.Error())
			continue
		}

		// If lock is acquired, release it immediately
		if err := locker.ReleaseLock(ctx, lockName, leaseID); err != nil {
			log.Printf("Failed to release temporary lock for table %s: %v", tableConfig.Name, err)
		} else {
			unlockedTables = append(unlockedTables, tableConfig)
		}
	}

	log.Printf("Unlocked tables: %v", unlockedTables)
	return unlockedTables
}

func (t *TableMonitoringService) ReleaseAllLocks(ctx context.Context) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for lockName, leaseID := range t.leaseIDs {
		log.Printf("Stopping lock renewal for %s", lockName)

		// Stop the renewal context for this lock
		if cancelFunc, exists := t.renewalCancelFuncs[lockName]; exists {
			cancelFunc() // Stop the renewal goroutine
			delete(t.renewalCancelFuncs, lockName)
		}

		// Create a locker for the table using the LockerFactory
		locker, err := t.lockerFactory.CreateLocker(lockName)
		if err != nil {
			log.Printf("Failed to create locker for lock %s during shutdown: %v", lockName, err)
			continue
		}

		// Attempt to release the lock using the correct lease ID
		if err := locker.ReleaseLock(ctx, lockName, leaseID); err != nil {
			log.Printf("Failed to release lock for %s: %v", lockName, err)
		} else {
			log.Printf("Lock released for %s", lockName)
		}
	}

	// Clear the lease IDs and cancellation functions after releasing all locks
	t.leaseIDs = map[string]string{}
	t.renewalCancelFuncs = map[string]context.CancelFunc{}
}
