package cdc

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/katasec/dstream/config"
)

// SQLServerMonitor manages SQL Server CDC monitoring for a given configuration
type SQLServerMonitor struct {
	dbConn   *sql.DB
	lastLSNs map[string][]byte
	lsnMutex sync.Mutex
}

// NewSQLServerMonitor initializes a new SQLServerMonitor with a database connection
func NewSQLServerMonitor(dbConn *sql.DB) *SQLServerMonitor {
	return &SQLServerMonitor{
		dbConn:   dbConn,
		lastLSNs: make(map[string][]byte),
	}
}

// MonitorTable continuously monitors a specific table for CDC changes
func (monitor *SQLServerMonitor) MonitorTable(tableConfig config.TableConfig) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in MonitorTable for table %s: %v", tableConfig.Name, r)
		}
	}()

	// Initialize LSN for this table
	err := monitor.initializeLSN(tableConfig.Name)
	if err != nil {
		return err
	}

	// Set the initial poll interval
	pollInterval, _ := tableConfig.GetPollInterval()

	for {
		// Fetch column names for the table
		columns, err := monitor.GetColumnNames("dbo", tableConfig.Name)
		if err != nil {
			log.Printf("Failed to get columns for %s: %v", tableConfig.Name, err)
			time.Sleep(pollInterval)
			continue
		}

		// Get the current LSN safely
		currentLSN := monitor.getCurrentLSN(tableConfig.Name)

		// Fetch CDC changes since the last known LSN
		changesFound, newLSN, err := fetchCDCChanges(monitor.dbConn, tableConfig.Name, columns, hex.EncodeToString(currentLSN))
		if err != nil {
			log.Printf("Error fetching changes for %s: %v", tableConfig.Name, err)
			time.Sleep(pollInterval)
			continue
		}

		// Update polling interval and LSN based on whether changes were found
		pollInterval, err = monitor.updatePollingIntervalAndLSN(tableConfig.Name, newLSN, changesFound, pollInterval, tableConfig)
		if err != nil {
			log.Printf("Error updating LSN and polling interval for %s: %v", tableConfig.Name, err)
		}

		// Log the next polling interval
		log.Printf("Next poll for table %s will occur in %s", tableConfig.Name, pollInterval)

		// Wait for the next poll interval
		time.Sleep(pollInterval)
	}
}

// initializeLSN loads the initial LSN for a table from the database into memory
func (monitor *SQLServerMonitor) initializeLSN(tableName string) error {
	defaultStartLSN := "00000000000000000000"
	initialLSN, err := loadLastLSN(monitor.dbConn, tableName, defaultStartLSN)
	if err != nil {
		log.Printf("Failed to load initial LSN for %s: %v", tableName, err)
		return err
	}

	// Store the initial LSN in the in-memory map
	monitor.lsnMutex.Lock()
	monitor.lastLSNs[tableName] = initialLSN
	monitor.lsnMutex.Unlock()

	return nil
}

// GetColumnNames retrieves column names for a given table in the schema
func (monitor *SQLServerMonitor) GetColumnNames(schema, tableName string) ([]string, error) {
	query := `SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = @schema AND TABLE_NAME = @tableName`
	rows, err := monitor.dbConn.Query(query, sql.Named("schema", schema), sql.Named("tableName", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return nil, err
		}
		columns = append(columns, columnName)
	}
	return columns, rows.Err()
}

// getCurrentLSN retrieves the current LSN for a table from the in-memory map with thread-safe locking
func (monitor *SQLServerMonitor) getCurrentLSN(tableName string) []byte {
	monitor.lsnMutex.Lock()
	defer monitor.lsnMutex.Unlock()
	return monitor.lastLSNs[tableName]
}

// updatePollingIntervalAndLSN adjusts the polling interval and updates the last LSN
func (monitor *SQLServerMonitor) updatePollingIntervalAndLSN(tableName string, newLSN []byte, changesFound bool, pollInterval time.Duration, tableConfig config.TableConfig) (time.Duration, error) {
	if changesFound {
		// Update the in-memory last LSN for this table
		monitor.lsnMutex.Lock()
		monitor.lastLSNs[tableName] = newLSN
		monitor.lsnMutex.Unlock()

		// Persist the new LSN to the database
		err := saveLastLSN(monitor.dbConn, tableName, newLSN)
		if err != nil {
			return pollInterval, fmt.Errorf("error saving last LSN for %s: %v", tableName, err)
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

	return pollInterval, nil
}
