package sqlserver

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/katasec/dstream/internal/cdc/utils"

	publishers "github.com/katasec/dstream/internal/publisher"
)



// SqlServerTableMonitor manages SQL Server CDC monitoring for a specific table
type SqlServerTableMonitor struct {
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

// NewSQLServerTableMonitor initializes a new SqlServerTableMonitor
func NewSQLServerTableMonitor(dbConn *sql.DB, tableName string, pollInterval, maxPollInterval time.Duration) *SqlServerTableMonitor {
	checkpointMgr := NewCheckpointManager(dbConn, tableName)

	// Fetch column names once and store them in the struct
	columns, err := fetchColumnNames(dbConn, tableName)
	if err != nil {
		log.Info("Failed to fetch column names", "table", tableName, "error", err)
		os.Exit(1)
	}

	return &SqlServerTableMonitor{
		dbConn:          dbConn,
		tableName:       tableName,
		pollInterval:    pollInterval,
		maxPollInterval: maxPollInterval,
		lastLSNs:        make(map[string][]byte),
		checkpointMgr:   checkpointMgr,
		columns:         columns,
	}
}

// MonitorTable continuously monitors the specified table
func (m *SqlServerTableMonitor) MonitorTable(ctx context.Context) error {
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
	backoff := utils.NewBackoffManager(m.pollInterval, m.maxPollInterval)

	// Begin monitoring loop
	for {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			log.Info("Stopping monitoring due to context cancellation", "table", m.tableName)
			return ctx.Err()
		default:
		}
		log.Info("Polling changes for table", "table", m.tableName, "lsn", hex.EncodeToString(m.lastLSNs[m.tableName]))
		changes, newLSN, err := m.fetchCDCChanges(m.lastLSNs[m.tableName])

		if err != nil {
			log.Info("Error fetching changes", "table", m.tableName, "error", err)
			time.Sleep(backoff.GetInterval()) // Wait with current interval on error
			continue
		}

		if len(changes) > 0 {
			log.Info("Changes detected, publishing...", "table", m.tableName)

			// Publish detected changes
			//consolePublisher := &publishers.ConsolePublisher{}
			for _, change := range changes {

				// Publish to console as a default
				// consolePublisher.PublishChange(change)

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
			log.Info("No changes found", "table", m.tableName, "nextPollIn", backoff.GetInterval())
		}

		time.Sleep(backoff.GetInterval())
	}
}

// fetchCDCChanges queries CDC changes and returns relevant events as a slice of maps
func (monitor *SqlServerTableMonitor) fetchCDCChanges(lastLSN []byte) ([]map[string]interface{}, []byte, error) {
	log.Info("Polling changes", "table", monitor.tableName, "lsn", hex.EncodeToString(lastLSN))

	// Use cached column names
	columnList := "ct.__$start_lsn, ct.__$operation"
	if len(monitor.columns) > 0 {
		columnList += ", " + strings.Join(monitor.columns, ", ")
	}
	query := fmt.Sprintf(`
        SELECT %s
        FROM TestDB.cdc.dbo_%s_CT AS ct
        WHERE ct.__$start_lsn > @lastLSN
        ORDER BY ct.__$start_lsn
    `, columnList, monitor.tableName)

	log.Info(query)
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

		// Determine the operation type as a string
		var operationType string
		switch operation {
		case 2:
			operationType = "Insert"
		case 4:
			operationType = "Update"
		case 1:
			operationType = "Delete"
		default:
			continue // Skip any unknown operation types
		}

		// Organize metadata and data separately in the output
		data := make(map[string]interface{})
		for i, colName := range monitor.columns {
			if colValue, ok := columnData[i+2].(*sql.NullString); ok && colValue.Valid {
				data[colName] = colValue.String
			} else {
				data[colName] = nil
			}
		}

		change := map[string]interface{}{
			"metadata": map[string]interface{}{
				"TableName":     monitor.tableName,
				"LSN":           hex.EncodeToString(lsn),
				"OperationID":   operation,
				"OperationType": operationType,
			},
			"data": data,
		}

		changes = append(changes, change)
		latestLSN = lsn
	}

	// Save the last LSN if changes were found
	if len(changes) > 0 {
		log.Info("Saving new last LSN for table %s: %x", monitor.tableName, latestLSN)
		err := monitor.checkpointMgr.SaveLastLSN(latestLSN)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to save last LSN for %s: %w", monitor.tableName, err)
		}
	}

	return changes, latestLSN, nil
}

// fetchColumnNames fetches column names for a specified table
func fetchColumnNames(db *sql.DB, tableName string) ([]string, error) {
	// Get the capture instance name for the table
	query := `
		SELECT COLUMN_NAME 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_NAME = @tableName 
		AND TABLE_SCHEMA = 'dbo'`

	log.Info(strings.Replace(query, "@tableName", "'"+tableName+"'", 1))

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
		log.Info("Found column", "name", columnName)
		columns = append(columns, columnName)
	}
	log.Info("Total columns found", "table", tableName, "columns", columns)
	return columns, rows.Err()
}
