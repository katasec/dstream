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

	"github.com/katasec/dstream/pkg/cdc"
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
	publisher       cdc.ChangePublisher
	columns         []string    // Cached column names
	batchSizer      *BatchSizer // Determines optimal batch size
}

// NewSQLServerTableMonitor initializes a new SqlServerTableMonitor
func NewSQLServerTableMonitor(dbConn *sql.DB, tableName string, pollInterval, maxPollInterval time.Duration, publisher cdc.ChangePublisher) *SqlServerTableMonitor {
	checkpointMgr := NewCheckpointManager(dbConn, tableName)

	// Fetch column names once and store them in the struct
	columns, err := fetchColumnNames(dbConn, tableName)
	if err != nil {
		log.Info("Failed to fetch column names", "table", tableName, "error", err)
		os.Exit(1)
	}

	// Create a BatchSizer with Standard SKU limit by default
	batchSizer := NewBatchSizer(dbConn, tableName, StandardSKULimit)

	return &SqlServerTableMonitor{
		dbConn:          dbConn,
		tableName:       tableName,
		pollInterval:    pollInterval,
		maxPollInterval: maxPollInterval,
		lastLSNs:        make(map[string][]byte),
		checkpointMgr:   checkpointMgr,
		columns:         columns,
		publisher:       publisher,
		batchSizer:      batchSizer,
	}
}

// MonitorTable continuously monitors the specified table
func (m *SqlServerTableMonitor) MonitorTable(ctx context.Context) error {
	err := m.checkpointMgr.InitializeCheckpointTable()
	if err != nil {
		return fmt.Errorf("error initializing checkpoint table: %w", err)
	}

	// Load last LSN for this table
	initialLSN, err := m.checkpointMgr.LoadLastLSN()
	if err != nil {
		return fmt.Errorf("error loading last LSN for table %s: %w", m.tableName, err)
	}

	// Multiple goroutines may access lastLSNs map, so lock it
	m.lsnMutex.Lock()
	m.lastLSNs[m.tableName] = initialLSN
	m.lsnMutex.Unlock()

	// Start the batch sizer to calculate optimal batch sizes
	err = m.batchSizer.Start(ctx)
	if err != nil {
		return fmt.Errorf("error starting batch sizer: %w", err)
	}

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
			log.Info("Changes detected, publishing...", "table", m.tableName, "changeCount", len(changes))

			// Publish all changes as a single batch
			doneChan, err := m.publisher.PublishChanges(changes)
			if err != nil {
				log.Error("Failed to publish batch", "error", err, "changeCount", len(changes))
				// Continue to next poll cycle without updating LSN
				time.Sleep(backoff.GetInterval())
				continue
			}

			// Wait for publish confirmation
			if success := <-doneChan; !success {
				log.Error("Failed to publish batch")
				// Continue to next poll cycle without updating LSN
				time.Sleep(backoff.GetInterval())
				continue
			}

			// Update last LSN and reset polling interval
			m.lsnMutex.Lock()
			m.lastLSNs[m.tableName] = newLSN
			m.lsnMutex.Unlock()

			// Update the checkpoint in the database
			if err := m.checkpointMgr.SaveLastLSN(newLSN); err != nil {
				log.Error("Failed to save checkpoint", "error", err)
			}

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

	// Get the optimal batch size from BatchSizer
	batchSize := monitor.batchSizer.GetBatchSize()
	log.Info("Using batch size", "table", monitor.tableName, "batchSize", batchSize)

	// Use cached column names
	columnList := "ct.__$start_lsn, ct.__$operation"
	if len(monitor.columns) > 0 {
		columnList += ", " + strings.Join(monitor.columns, ", ")
	}

	query := fmt.Sprintf(`
        SELECT TOP(%d) %s
        FROM cdc.dbo_%s_CT AS ct
        WHERE ct.__$start_lsn > @lastLSN
        ORDER BY ct.__$start_lsn
    `, batchSize, columnList, monitor.tableName)

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

	// Return the changes and latest LSN without saving the checkpoint yet
	// The checkpoint will be saved in MonitorTable after successful publishing

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
