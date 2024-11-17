package cdc

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/katasec/dstream/cdc/publishers"
	"github.com/katasec/dstream/config"
)

// TableMonitoringService manages monitoring for each table in the config.
type TableMonitoringService struct {
	db            *sql.DB
	config        *config.Config
	lockerFactory *LockerFactory
}

// NewTableMonitoringService initializes a new TableMonitoringService.
func NewTableMonitoringService(db *sql.DB, config *config.Config) *TableMonitoringService {
	return &TableMonitoringService{
		db:            db,
		config:        config,
		lockerFactory: NewLockerFactory(config),
	}
}

// StartMonitoring initializes and starts monitoring for each table in the config.
func (t *TableMonitoringService) StartMonitoring() error {
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

		// Create a locker for the table using the LockerFactory
		locker, err := t.lockerFactory.CreateLocker(lockName)
		if err != nil {
			log.Printf("Failed to create locker for table %s: %v", tableConfig.Name, err)
			wg.Done()
			continue
		}

		// Acquire the lock
		leaseID, err := locker.AcquireLock(context.TODO(), lockName)
		if leaseID == "" {
			log.Printf("Skipping table %s as it is already locked.", tableConfig.Name)
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
			locker.StartLockRenewal(context.TODO(), lockName)

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
		_, err = locker.AcquireLock(ctx, lockName)
		if err != nil {
			log.Printf("Table %s is already locked. Skipping...", tableConfig.Name)
			continue
		}

		// If lock is acquired, release it immediately
		if err := locker.ReleaseLock(ctx, lockName); err != nil {
			log.Printf("Failed to release temporary lock for table %s: %v", tableConfig.Name, err)
		} else {
			unlockedTables = append(unlockedTables, tableConfig)
		}
	}

	log.Printf("Unlocked tables: %v", unlockedTables)
	return unlockedTables
}
