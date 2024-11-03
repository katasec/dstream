package cdc

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
)

// InitializeCheckpointTable checks for the existence of the cdc_offsets table and creates it if it doesn't exist
func InitializeCheckpointTable(db *sql.DB) error {
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

// Load the last LSN from the database for a specific table
func loadLastLSN(db *sql.DB, tableName, defaultStartLSN string) ([]byte, error) {
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
func saveLastLSN(db *sql.DB, tableName string, newLSN []byte) error {
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
