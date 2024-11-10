package cdc

import (
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/katasec/dstream/config"
)

// TableMonitoringService manages monitoring for each table in the config.
type TableMonitoringService struct {
	db     *sql.DB
	config *config.Config
}

// NewTableMonitoringService initializes a new TableMonitoringService.
func NewTableMonitoringService(db *sql.DB, config *config.Config) *TableMonitoringService {
	return &TableMonitoringService{
		db:     db,
		config: config,
	}
}

// StartMonitoring initializes and starts monitoring for each table in the config.
func (t *TableMonitoringService) StartMonitoring() error {
	var wg sync.WaitGroup // WaitGroup to ensure goroutines complete

	// Initialize ChangePublisherFactory
	publisherFactory := NewChangePublisherFactory(t.config)

	for _, tableConfig := range t.config.Tables {
		wg.Add(1) // Increment the WaitGroup counter for each table

		pollInterval, _ := tableConfig.GetPollInterval()
		maxPollInterval, _ := tableConfig.GetMaxPollInterval()

		// Create the appropriate publisher for the table
		publisher, err := publisherFactory.Create()
		if err != nil {
			log.Printf("Error creating publisher for table %s: %v", tableConfig.Name, err)
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
		go func(monitor *SQLServerMonitor, tableConfig config.TableConfig) {
			defer wg.Done() // Mark goroutine as done when it completes

			log.Printf("Starting monitor for table: %s", tableConfig.Name)
			if err := monitor.MonitorTable(); err != nil {
				log.Printf("Error monitoring table %s: %v", tableConfig.Name, err)
			} else {
				log.Printf("Monitoring completed for table %s", tableConfig.Name)
			}
		}(monitor, tableConfig)

		// Stagger the start of each monitor by a short interval
		time.Sleep(500 * time.Millisecond)
	}

	// Wait for all monitoring goroutines to complete
	wg.Wait()
	log.Println("All table monitors have completed.")
	return nil
}