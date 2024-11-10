// checkpoint_manager.go
package cdc

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
)

// CheckpointManager manages checkpoint (LSN) persistence
type CheckpointManager struct {
	dbConn    *sql.DB
	tableName string
}

// NewCheckpointManager initializes a new CheckpointManager
func NewCheckpointManager(dbConn *sql.DB, tableName string) *CheckpointManager {
	return &CheckpointManager{
		dbConn:    dbConn,
		tableName: tableName,
	}
}

// InitializeCheckpointTable creates the checkpoint table if it does not exist
func (c *CheckpointManager) InitializeCheckpointTable() error {
	query := `
    IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'cdc_offsets')
    BEGIN
        CREATE TABLE cdc_offsets (
            table_name NVARCHAR(255) PRIMARY KEY,
            last_lsn VARBINARY(10),
            updated_at DATETIME DEFAULT GETDATE()
        );
    END`

	_, err := c.dbConn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create cdc_offsets table: %w", err)
	}

	log.Println("Initialized checkpoints table.")
	return nil
}

// LoadLastLSN retrieves the last known LSN for the specified table
func (c *CheckpointManager) LoadLastLSN(defaultStartLSN string) ([]byte, error) {
	var lastLSN []byte
	query := "SELECT last_lsn FROM cdc_offsets WHERE table_name = @tableName"
	err := c.dbConn.QueryRow(query, sql.Named("tableName", c.tableName)).Scan(&lastLSN)
	if err == sql.ErrNoRows {
		startLSNBytes, _ := hex.DecodeString(defaultStartLSN)
		log.Printf("No previous LSN for %s. Initializing with default start LSN.", c.tableName)
		return startLSNBytes, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to load LSN for %s: %w", c.tableName, err)
	}
	log.Printf("Resuming %s from last LSN: %s", c.tableName, hex.EncodeToString(lastLSN))
	return lastLSN, nil
}

// SaveLastLSN updates the last known LSN for the specified table
func (c *CheckpointManager) SaveLastLSN(newLSN []byte) error {
	upsertQuery := `
    MERGE INTO cdc_offsets AS target
    USING (VALUES (@tableName, @lastLSN, GETDATE())) AS source (table_name, last_lsn, updated_at)
    ON target.table_name = source.table_name
    WHEN MATCHED THEN 
        UPDATE SET last_lsn = source.last_lsn, updated_at = source.updated_at
    WHEN NOT MATCHED THEN
        INSERT (table_name, last_lsn, updated_at) 
        VALUES (source.table_name, source.last_lsn, source.updated_at);`

	_, err := c.dbConn.Exec(upsertQuery, sql.Named("tableName", c.tableName), sql.Named("lastLSN", newLSN))
	if err != nil {
		return fmt.Errorf("failed to save LSN for %s: %w", c.tableName, err)
	}

	log.Printf("Saved new LSN for %s: %s", c.tableName, hex.EncodeToString(newLSN))
	return nil
}
