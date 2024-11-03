package cdc

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// JsonEvent represents the simplified JSON structure for CDC events
type JsonEvent struct {
	Table string                 `json:"table"`
	Op    string                 `json:"op"`
	Data  map[string]interface{} `json:"data,omitempty"`
	LSN   string                 `json:"lsn"`
}

// fetchCDCChanges processes CDC changes and converts them to JSON events
func fetchCDCChanges(db *sql.DB, tableName string, columns []string, defaultStartLSN string) (bool, []byte, error) {
	changesFound := false
	var latestLSN []byte

	lastLSN, err := loadLastLSN(db, tableName, defaultStartLSN)
	if err != nil {
		return false, nil, fmt.Errorf("failed to load last LSN for %s: %w", tableName, err)
	}

	rows, err := getCDCEntries(db, tableName, columns, lastLSN)
	if err != nil {
		return false, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		lsn, operation, data, err := scanRowData(rows, columns)
		if err != nil {
			log.Printf("Error scanning row for table %s: %v", tableName, err)
			continue
		}

		if err := processCdcEvent(tableName, operation, data, lsn); err != nil {
			log.Printf("Error processing event for table %s: %v", tableName, err)
			continue
		}

		changesFound = true
		latestLSN = lsn
	}

	return changesFound, latestLSN, nil
}

// getCDCEntries retrieves CDC entries for a given table and last LSN
func getCDCEntries(db *sql.DB, tableName string, columns []string, lastLSN []byte) (*sql.Rows, error) {
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
func scanRowData(rows *sql.Rows, columns []string) ([]byte, int, map[string]interface{}, error) {
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

func processCdcEvent(tableName string, operation int, data map[string]interface{}, lsn []byte) error {
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
		return nil
	}

	jsonData, err := toJsonEvent(tableName, op, data, lsn)
	if err != nil {
		return fmt.Errorf("error generating JSON event: %w", err)
	}

	log.Println(string(jsonData))
	return nil
}

// toJsonEvent formats the CDC event as a JsonEvent instance
func toJsonEvent(
	tableName string,
	operation string,
	data map[string]interface{},
	lsn []byte,
) ([]byte, error) {
	event := JsonEvent{
		Table: tableName,
		Op:    operation,
		Data:  data,
		LSN:   hex.EncodeToString(lsn),
	}

	return json.MarshalIndent(event, "", "    ")
}
