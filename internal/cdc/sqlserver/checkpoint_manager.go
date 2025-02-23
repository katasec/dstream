package sqlserver

import (
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/katasec/dstream/internal/logging"
)

var log = logging.GetLogger()

// Default checkpoint table name
const defaultCheckpointTableName = "cdc_offsets"

// CheckpointManager manages checkpoint (LSN) persistence
type CheckpointManager struct {
	dbConn          *sql.DB
	tableName       string
	checkpointTable string
}

// NewCheckpointManager initializes a new CheckpointManager
func NewCheckpointManager(dbConn *sql.DB, tableName string, checkpointTableName ...string) *CheckpointManager {
	// Use provided checkpoint table name if supplied; otherwise, use default
	cpTable := defaultCheckpointTableName
	if len(checkpointTableName) > 0 && checkpointTableName[0] != "" {
		cpTable = checkpointTableName[0]
	}

	return &CheckpointManager{
		dbConn:          dbConn,
		tableName:       tableName,
		checkpointTable: cpTable,
	}
}

// InitializeCheckpointTable creates the checkpoint table if it does not exist
func (c *CheckpointManager) InitializeCheckpointTable() error {
	query := fmt.Sprintf(`
    IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = '%s')
    BEGIN
        CREATE TABLE %s (
            table_name NVARCHAR(255) PRIMARY KEY,
            last_lsn VARBINARY(10),
            updated_at DATETIME DEFAULT GETDATE()
        );
    END`, c.checkpointTable, c.checkpointTable)

	_, err := c.dbConn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create %s table: %w", c.checkpointTable, err)
	}

	log.Info("Initialized checkpoints table")
	return nil
}

// LoadLastLSN retrieves the last known LSN for the specified table
func (c *CheckpointManager) LoadLastLSN(defaultStartLSN string) ([]byte, error) {
	var lastLSN []byte
	query := fmt.Sprintf("SELECT last_lsn FROM %s WHERE table_name = @tableName", c.checkpointTable)
	err := c.dbConn.QueryRow(query, sql.Named("tableName", c.tableName)).Scan(&lastLSN)
	if err == sql.ErrNoRows {
		startLSNBytes, _ := hex.DecodeString(defaultStartLSN)
		log.Info("No previous LSN, initializing with default", "table", c.tableName)
		return startLSNBytes, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to load LSN for %s: %w", c.tableName, err)
	}
	log.Info("Resuming from last LSN", "table", c.tableName, "lsn", hex.EncodeToString(lastLSN))
	return lastLSN, nil
}

// SaveLastLSN updates the last known LSN for the specified table
func (c *CheckpointManager) SaveLastLSN(newLSN []byte) error {
	upsertQuery := fmt.Sprintf(`
    MERGE INTO %s AS target
    USING (VALUES (@tableName, @lastLSN, GETDATE())) AS source (table_name, last_lsn, updated_at)
    ON target.table_name = source.table_name
    WHEN MATCHED THEN 
        UPDATE SET last_lsn = source.last_lsn, updated_at = source.updated_at
    WHEN NOT MATCHED THEN
        INSERT (table_name, last_lsn, updated_at) 
        VALUES (source.table_name, source.last_lsn, source.updated_at);`, c.checkpointTable)

	_, err := c.dbConn.Exec(upsertQuery, sql.Named("tableName", c.tableName), sql.Named("lastLSN", newLSN))
	if err != nil {
		return fmt.Errorf("failed to save LSN for %s: %w", c.tableName, err)
	}

	log.Info("Saved new LSN", "table", c.tableName, "lsn", hex.EncodeToString(newLSN))
	return nil
}
