package cdc

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	publishers "github.com/katasec/dstream/cdc/publishers"
)

// SQLServerMonitor manages SQL Server CDC monitoring for a specific table
type SQLServerMonitor struct {
	dbConn          *sql.DB
	tableName       string
	pollInterval    time.Duration
	maxPollInterval time.Duration
	lastLSNs        map[string][]byte
	lsnMutex        sync.Mutex
	checkpointMgr   *CheckpointManager
	publisher       publishers.ChangePublisher
	columns         []string // Cached column names
}

// NewSQLServerMonitor initializes a new SQLServerMonitor for a specific table
func NewSQLServerMonitor(dbConn *sql.DB, tableName string, pollInterval, maxPollInterval time.Duration, publisher publishers.ChangePublisher) *SQLServerMonitor {
	checkpointMgr := NewCheckpointManager(dbConn, tableName)

	// Fetch column names once and store them in the struct
	columns, err := fetchColumnNames(dbConn, tableName)
	if err != nil {
		log.Fatalf("Failed to fetch column names for table %s: %v", tableName, err)
	}

	return &SQLServerMonitor{
		dbConn:          dbConn,
		tableName:       tableName,
		pollInterval:    pollInterval,
		maxPollInterval: maxPollInterval,
		lastLSNs:        make(map[string][]byte),
		checkpointMgr:   checkpointMgr,
		publisher:       publisher,
		columns:         columns,
	}
}

// MonitorTable continuously monitors the specified table
func (m *SQLServerMonitor) MonitorTable() error {
	err := m.checkpointMgr.InitializeCheckpointTable()
	if err != nil {
		return fmt.Errorf("error initializing checkpoint table: %w", err)
	}

	// Load last LSN for this table
	defaultStartLSN := "00000000000000000000"
	initialLSN, err := m.checkpointMgr.LoadLastLSN(defaultStartLSN)
	if err != nil {
		return fmt.Errorf("error loading last LSN for table %s: %w", m.tableName, err)
	}
	m.lsnMutex.Lock()
	m.lastLSNs[m.tableName] = initialLSN
	m.lsnMutex.Unlock()

	// Initialize the backoff manager
	backoff := NewBackoffManager(m.pollInterval, m.maxPollInterval)

	// Begin monitoring loop
	for {
		log.Printf("Polling changes for table: %s", m.tableName)
		changes, newLSN, err := m.fetchCDCChanges(m.lastLSNs[m.tableName])

		if err != nil {
			log.Printf("Error fetching changes for %s: %v", m.tableName, err)
			time.Sleep(backoff.GetInterval()) // Wait with current interval on error
			continue
		}

		if len(changes) > 0 {
			log.Printf("Changes detected for table %s; publishing...", m.tableName)

			// Publish detected changes
			consolePublisher := &publishers.ConsolePublisher{}
			for _, change := range changes {

				// Publish to console as a default
				consolePublisher.PublishChange(change)

				// publish to publisher configured for this table
				m.publisher.PublishChange(change)
			}

			// Update last LSN and reset polling interval
			m.lsnMutex.Lock()
			m.lastLSNs[m.tableName] = newLSN
			m.lsnMutex.Unlock()
			backoff.ResetInterval() // Reset interval after detecting changes

		} else {
			// If no changes, increase the polling interval (backoff)
			backoff.IncreaseInterval()
			log.Printf("No changes found for table %s. Next poll in %s", m.tableName, backoff.GetInterval())
		}

		time.Sleep(backoff.GetInterval())
	}
}

// fetchCDCChanges queries CDC changes and returns them as a slice of maps
func (monitor *SQLServerMonitor) fetchCDCChanges(lastLSN []byte) ([]map[string]interface{}, []byte, error) {
	log.Printf("Polling changes for table: %s with last LSN: %x", monitor.tableName, lastLSN)

	// Use cached column names
	columnList := "ct.__$start_lsn, ct.__$operation, " + strings.Join(monitor.columns, ", ")
	query := fmt.Sprintf(`
        SELECT %s
        FROM cdc.dbo_%s_CT AS ct
        WHERE ct.__$start_lsn > @lastLSN
        ORDER BY ct.__$start_lsn
    `, columnList, monitor.tableName)

	rows, err := monitor.dbConn.Query(query, sql.Named("lastLSN", lastLSN))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query CDC table for %s: %w", monitor.tableName, err)
	}
	defer rows.Close()

	changes := []map[string]interface{}{}
	var latestLSN []byte

	for rows.Next() {
		var lsn []byte
		var operation int
		columnData := make([]interface{}, len(monitor.columns)+2)
		columnData[0] = &lsn
		columnData[1] = &operation
		for i := range monitor.columns {
			var colValue sql.NullString
			columnData[i+2] = &colValue
		}

		if err := rows.Scan(columnData...); err != nil {
			return nil, nil, fmt.Errorf("failed to scan row: %w", err)
		}

		data := map[string]interface{}{"LSN": hex.EncodeToString(lsn), "Operation": operation}
		for i, colName := range monitor.columns {
			if colValue, ok := columnData[i+2].(*sql.NullString); ok && colValue.Valid {
				data[colName] = colValue.String
			} else {
				data[colName] = nil
			}
		}

		changes = append(changes, data)
		latestLSN = lsn
	}

	// Save the last LSN if changes were found
	if len(changes) > 0 {
		log.Printf("Saving new last LSN for table %s: %x", monitor.tableName, latestLSN)
		err := monitor.checkpointMgr.SaveLastLSN(latestLSN)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to save last LSN for %s: %w", monitor.tableName, err)
		}
	}

	return changes, latestLSN, nil
}

// fetchColumnNames fetches column names for a specified table
func fetchColumnNames(db *sql.DB, tableName string) ([]string, error) {
	query := `SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = @tableName`
	rows, err := db.Query(query, sql.Named("tableName", tableName))
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
