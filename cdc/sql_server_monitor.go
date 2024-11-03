package cdc

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
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

// InitializeCheckpointTable checks for the existence of the cdc_offsets table and creates it if it doesn't exist
func (monitor *SQLServerMonitor) InitializeCheckpointTable(db *sql.DB) error {
	query := `
	IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'cdc_offsets')
	BEGIN
		CREATE TABLE cdc_offsets (
			table_name NVARCHAR(255) PRIMARY KEY,
			last_lsn VARBINARY(10),
			updated_at DATETIME DEFAULT GETDATE()
		);
	END`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create cdc_offsets table: %w", err)
	}

	log.Println("cdc_offsets table is ready (created if it didn't exist).")
	return nil
}

// StartMonitoring launches monitoring for each configured table based on the SQLServerMonitor struct
func (monitor *SQLServerMonitor) StartMonitoring(dbConn *sql.DB, cfg config.Config) {
	sqlMonitor := NewSQLServerMonitor(dbConn)
	for _, tableConfig := range cfg.Tables {
		go func(tableConfig config.TableConfig) {
			err := sqlMonitor.MonitorTable(tableConfig)
			if err != nil {
				log.Printf("Monitoring stopped for table %s due to error: %v", tableConfig.Name, err)
			}
		}(tableConfig)
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
		changesFound, newLSN, err := monitor.fetchCDCChanges(monitor.dbConn, tableConfig.Name, columns, hex.EncodeToString(currentLSN))
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
	initialLSN, err := monitor.loadLastLSN(monitor.dbConn, tableName, defaultStartLSN)
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
func (m *SQLServerMonitor) updatePollingIntervalAndLSN(tableName string, newLSN []byte, changesFound bool, pollInterval time.Duration, tableConfig config.TableConfig) (time.Duration, error) {
	if changesFound {
		// Update the in-memory last LSN for this table
		m.lsnMutex.Lock()
		m.lastLSNs[tableName] = newLSN
		m.lsnMutex.Unlock()

		// Persist the new LSN to the database
		err := m.saveLastLSN(m.dbConn, tableName, newLSN)
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

// fetchCDCChanges processes CDC changes and converts them to JSON events
func (m *SQLServerMonitor) fetchCDCChanges(db *sql.DB, tableName string, columns []string, defaultStartLSN string) (bool, []byte, error) {
	changesFound := false
	var latestLSN []byte

	lastLSN, err := m.loadLastLSN(db, tableName, defaultStartLSN)
	if err != nil {
		return false, nil, fmt.Errorf("failed to load last LSN for %s: %w", tableName, err)
	}

	rows, err := m.getCDCEntries(db, tableName, columns, lastLSN)
	if err != nil {
		return false, nil, err
	}
	defer rows.Close()

	for rows.Next() {

		// Read Operation from CDC tables
		lsn, operation, data, err := m.scanRowData(rows, columns)
		if err != nil {
			log.Printf("Error scanning row for table %s: %v", tableName, err)
			continue
		}

		// Generate JSON payload
		var jsonData string
		jsonData, err = m.genJsonForEvent(tableName, operation, data, lsn)
		if err != nil {
			log.Printf("Error processing event for table %s: %v", tableName, err)
			continue
		}
		log.Println(jsonData)

		changesFound = true
		latestLSN = lsn
	}

	return changesFound, latestLSN, nil
}

// getCDCEntries retrieves CDC entries for a given table and last LSN
func (monitor *SQLServerMonitor) getCDCEntries(db *sql.DB, tableName string, columns []string, lastLSN []byte) (*sql.Rows, error) {
	columnList := "__$start_lsn, __$operation, " + strings.Join(columns, ", ")
	query := fmt.Sprintf(`
        SELECT %s
        FROM cdc.dbo_%s_CT
        WHERE __$start_lsn > @lastLSN
        ORDER BY __$start_lsn
    `, columnList, tableName)

	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement for %s: %w", tableName, err)
	}

	rows, err := stmt.Query(sql.Named("lastLSN", lastLSN))
	if err != nil {
		stmt.Close()
		return nil, fmt.Errorf("failed to query CDC table for %s: %w", tableName, err)
	}
	return rows, nil
}

// scanRowData scans a single row from the CDC result set into variables
func (m *SQLServerMonitor) scanRowData(rows *sql.Rows, columns []string) ([]byte, int, map[string]interface{}, error) {
	var lsn []byte
	var operation int

	scanTargets := make([]interface{}, len(columns)+2)
	scanTargets[0] = &lsn
	scanTargets[1] = &operation

	data := make(map[string]interface{})
	for i := range columns {
		var colValue sql.NullString
		scanTargets[i+2] = &colValue
	}

	if err := rows.Scan(scanTargets...); err != nil {
		return nil, 0, nil, fmt.Errorf("failed to scan row: %w", err)
	}

	for i, colName := range columns {
		if colValue, ok := scanTargets[i+2].(*sql.NullString); ok && colValue.Valid {
			data[colName] = colValue.String
		} else {
			data[colName] = nil
		}
	}

	return lsn, operation, data, nil
}

func (m *SQLServerMonitor) genJsonForEvent(tableName string, operation int, data map[string]interface{}, lsn []byte) (string, error) {
	var op string
	switch operation {
	case 1:
		op = "delete"
	case 2:
		op = "insert"
	case 4: // Only capture the "after" image of the update
		op = "update"
	default:
		// Ignore "before image" for updates (operation = 3)
		return "", nil
	}

	jsonEvent := &JsonEvent{
		Table: tableName,
		Op:    op,
		Data:  data,
		LSN:   hex.EncodeToString(lsn),
	}
	jsonData, err := jsonEvent.toJsonPayload()
	if err != nil {
		return "", fmt.Errorf("error generating JSON event: %w", err)
	}

	return string(jsonData), nil
}

// Load the last LSN from the database for a specific table
func (m *SQLServerMonitor) loadLastLSN(db *sql.DB, tableName, defaultStartLSN string) ([]byte, error) {
	var lastLSN []byte
	query := "SELECT last_lsn FROM cdc_offsets WHERE table_name = @tableName"
	err := db.QueryRow(query, sql.Named("tableName", tableName)).Scan(&lastLSN)
	if err == sql.ErrNoRows {
		// If no checkpoint exists, initialize with a default LSN
		startLSNBytes, _ := hex.DecodeString(defaultStartLSN)
		lastLSN = startLSNBytes
		log.Printf("No previous LSN for %s. Initializing with default start LSN.", tableName)
	} else if err != nil {
		return nil, fmt.Errorf("failed to load LSN for %s: %w", tableName, err)
	} else {
		log.Printf("Resuming %s from last LSN: %s", tableName, hex.EncodeToString(lastLSN))
	}
	return lastLSN, nil
}

// Save the last processed LSN for a specific table to the database if it has changed
func (m *SQLServerMonitor) saveLastLSN(db *sql.DB, tableName string, newLSN []byte) error {
	// Check if the last LSN in the database matches the new LSN
	var currentLSN []byte
	query := "SELECT last_lsn FROM cdc_offsets WHERE table_name = @tableName"
	err := db.QueryRow(query, sql.Named("tableName", tableName)).Scan(&currentLSN)

	if err == nil && string(currentLSN) == string(newLSN) {
		// If the current LSN matches the new LSN, skip updating
		log.Printf("No change in LSN for %s; skipping save.", tableName)
		return nil
	} else if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check current LSN for %s: %w", tableName, err)
	}

	// SQL Server-compatible upsert using MERGE
	upsertQuery := `
	MERGE INTO cdc_offsets AS target
	USING (VALUES (@tableName, @lastLSN, GETDATE())) AS source (table_name, last_lsn, updated_at)
	ON target.table_name = source.table_name
	WHEN MATCHED THEN 
		UPDATE SET last_lsn = source.last_lsn, updated_at = source.updated_at
	WHEN NOT MATCHED THEN
		INSERT (table_name, last_lsn, updated_at) 
		VALUES (source.table_name, source.last_lsn, source.updated_at);`

	_, err = db.Exec(upsertQuery, sql.Named("tableName", tableName), sql.Named("lastLSN", newLSN))
	if err != nil {
		return fmt.Errorf("failed to save LSN for %s: %w", tableName, err)
	}

	log.Printf("Saved new LSN for %s: %s", tableName, hex.EncodeToString(newLSN))
	return nil
}
