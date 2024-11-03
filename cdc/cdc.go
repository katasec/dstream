package cdc

import (
	"database/sql"
	"encoding/hex"
	"log"
	"sync"
	"time"

	"github.com/katasec/dstream/config"
	"github.com/katasec/dstream/db"
)

var (
	lastLSNs = make(map[string][]byte) // In-memory map to store last LSNs per table
	lsnMutex = sync.Mutex{}            // Mutex to synchronize access to lastLSNs map
)

// StartMonitoring launches a monitoring goroutine for each configured table
func StartMonitoring(dbConn *sql.DB, cfg config.Config) {
	for _, tableConfig := range cfg.Tables {
		go func(tableConfig config.TableConfig) {
			err := monitorTable(dbConn, tableConfig)
			if err != nil {
				log.Printf("Monitoring stopped for table %s due to error: %v", tableConfig.Name, err)
			}
		}(tableConfig)
	}
}

// monitorTable continuously monitors a specific table for CDC changes
func monitorTable(dbConn *sql.DB, tableConfig config.TableConfig) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in monitorTable for table %s: %v", tableConfig.Name, r)
		}
	}()

	// Load initial LSN for this table from the database into memory
	defaultStartLSN := "00000000000000000000"
	initialLSN, err := loadLastLSN(dbConn, tableConfig.Name, defaultStartLSN)
	if err != nil {
		log.Printf("Failed to load initial LSN for %s: %v", tableConfig.Name, err)
		return err
	}

	// Store the initial LSN in the in-memory map
	lsnMutex.Lock()
	lastLSNs[tableConfig.Name] = initialLSN
	lsnMutex.Unlock()

	// Set the initial poll interval
	pollInterval, _ := tableConfig.GetPollInterval()

	for {
		// Fetch column names for the table
		columns, err := db.GetColumnNames(dbConn, "dbo", tableConfig.Name)
		if err != nil {
			log.Printf("Failed to get columns for %s: %v", tableConfig.Name, err)
			time.Sleep(pollInterval)
			continue
		}

		// Lock to safely read the current LSN from the map
		lsnMutex.Lock()
		currentLSN := lastLSNs[tableConfig.Name]
		lsnMutex.Unlock()

		// Fetch CDC changes since the last known LSN
		changesFound, newLSN, err := fetchCDCChanges(dbConn, tableConfig.Name, columns, hex.EncodeToString(currentLSN))
		if err != nil {
			log.Printf("Error fetching changes for %s: %v", tableConfig.Name, err)
			time.Sleep(pollInterval)
			continue
		}

		// If changes are found, update the last LSN
		if changesFound {
			// Update the in-memory last LSN for this table
			lsnMutex.Lock()
			lastLSNs[tableConfig.Name] = newLSN
			lsnMutex.Unlock()

			// Persist the new LSN to the database
			err := saveLastLSN(dbConn, tableConfig.Name, newLSN)
			if err != nil {
				log.Printf("Error saving last LSN for %s: %v", tableConfig.Name, err)
			}

			// Reset poll interval after successful processing
			pollInterval, _ = tableConfig.GetPollInterval()
		} else {
			// Back off the poll interval up to the max interval if no changes were found
			pollInterval *= 2
			maxPollInterval, _ := tableConfig.GetMaxPollInterval()
			if pollInterval > maxPollInterval {
				pollInterval = maxPollInterval
			}
		}

		// Log the next polling interval
		log.Printf("Next poll for table %s will occur in %s", tableConfig.Name, pollInterval)

		// Wait for the next poll interval
		time.Sleep(pollInterval)
	}
}
