package cdc

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

// toJsonEvent converts raw CDC data to a JSON event compatible with Debezium’s structure
func toJsonEvent(
	tableName string,
	operation int,
	before, after map[string]interface{},
	lsn []byte,
) ([]byte, error) {
	// Map operation types to JSON-compatible op codes
	var op string
	switch operation {
	case 1:
		op = "c" // Create
	case 2:
		op = "u" // Update
	case 3:
		op = "d" // Delete
	default:
		return nil, fmt.Errorf("unsupported operation type: %d", operation)
	}

	// Prepare JSON event structure
	event := struct {
		Schema  interface{} `json:"schema"`
		Payload struct {
			Before map[string]interface{} `json:"before,omitempty"`
			After  map[string]interface{} `json:"after,omitempty"`
			Source map[string]interface{} `json:"source"`
			Op     string                 `json:"op"`
			TsMs   int64                  `json:"ts_ms"`
		} `json:"payload"`
	}{
		Schema: nil,
		Payload: struct {
			Before map[string]interface{} `json:"before,omitempty"`
			After  map[string]interface{} `json:"after,omitempty"`
			Source map[string]interface{} `json:"source"`
			Op     string                 `json:"op"`
			TsMs   int64                  `json:"ts_ms"`
		}{
			Before: before,
			After:  after,
			Source: map[string]interface{}{
				"version":   "1.9.5.Final",
				"connector": "sqlserver",
				"name":      "dbserver1",
				"db":        "TestDB",
				"schema":    "dbo",
				"table":     tableName,
				"txId":      "example_tx_id", // Placeholder; replace as needed
				"lsn":       hex.EncodeToString(lsn),
				"commit":    true,
			},
			Op:   op,
			TsMs: time.Now().UnixMilli(),
		},
	}

	// Marshal the event to JSON
	return json.MarshalIndent(event, "", "    ")
}

// fetchCDCChanges adapted to call `toJsonEvent` for output formatting
func fetchCDCChanges(db *sql.DB, tableName string, columns []string, defaultStartLSN string) (bool, []byte, error) {
	changesFound := false
	var latestLSN []byte

	// Load the last LSN from the checkpoint table
	lastLSN, err := loadLastLSN(db, tableName, defaultStartLSN)
	if err != nil {
		return false, nil, fmt.Errorf("failed to load last LSN for %s: %w", tableName, err)
	}

	// Prepare query
	columnList := "__$start_lsn, __$operation, " + strings.Join(columns, ", ")
	query := fmt.Sprintf(`
        SELECT %s
        FROM cdc.dbo_%s_CT
        WHERE __$start_lsn > @lastLSN
        ORDER BY __$start_lsn
    `, columnList, tableName)

	stmt, err := db.Prepare(query)
	if err != nil {
		return false, nil, fmt.Errorf("failed to prepare statement for %s: %w", tableName, err)
	}
	defer stmt.Close()

	// Query with last LSN parameter
	rows, err := stmt.Query(sql.Named("lastLSN", lastLSN))
	if err != nil {
		return false, nil, fmt.Errorf("failed to query CDC table for %s: %w", tableName, err)
	}
	defer rows.Close()

	// Process each row
	for rows.Next() {
		lsn := []byte{}
		operation := 0
		scanTargets := make([]interface{}, len(columns)+2)
		scanTargets[0] = &lsn
		scanTargets[1] = &operation

		before := make(map[string]interface{})
		after := make(map[string]interface{})
		for i := range columns {
			var colValue sql.NullString
			scanTargets[i+2] = &colValue
		}

		// Scan the row data into scanTargets
		if err := rows.Scan(scanTargets...); err != nil {
			return false, nil, fmt.Errorf("failed to scan row for %s: %w", tableName, err)
		}

		// Populate the "after" map with current column values
		for i, colName := range columns {
			if colValue, ok := scanTargets[i+2].(*sql.NullString); ok && colValue.Valid {
				after[colName] = colValue.String
			} else {
				after[colName] = nil
			}
		}

		// Set `before` only for update and delete operations
		if operation == 2 || operation == 3 {
			before = after // In this example, we’re using the same data for simplicity
		}

		// Generate JSON-compatible event
		jsonData, err := toJsonEvent(tableName, operation, before, after, lsn)
		if err != nil {
			log.Printf("Error generating JSON event for table %s: %v", tableName, err)
			continue
		}

		// Log the event
		log.Println(string(jsonData))

		// Update latestLSN to current row's LSN
		changesFound = true
		latestLSN = lsn
	}

	return changesFound, latestLSN, nil
}
