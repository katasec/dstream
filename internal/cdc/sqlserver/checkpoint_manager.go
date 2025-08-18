package sqlserver

import (
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/katasec/dstream/internal/logging"
)

var log = logging.GetLogger()
var defaultStartLSN = "00000000000000000000"
var defaultStartSEQ = "00000000000000000000"

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
	// Create table if not exists
	createQuery := fmt.Sprintf(`
	IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = '%s')
	BEGIN
		CREATE TABLE %s (
			table_name NVARCHAR(255) PRIMARY KEY,
			last_lsn VARBINARY(10),
			updated_at DATETIME DEFAULT GETDATE(),
			last_seq VARBINARY(10)
		);
	END`, c.checkpointTable, c.checkpointTable)

	_, err := c.dbConn.Exec(createQuery)
	if err != nil {
		return fmt.Errorf("failed to create %s table: %w", c.checkpointTable, err)
	}

	// Check if last_seq column exists
	checkColQuery := fmt.Sprintf(`
	SELECT COUNT(*) FROM sys.columns 
	WHERE Name = N'last_seq' AND Object_ID = Object_ID(N'%s')`, c.checkpointTable)
	var colCount int
	err = c.dbConn.QueryRow(checkColQuery).Scan(&colCount)
	if err != nil {
		return fmt.Errorf("failed to check last_seq column: %w", err)
	}

	// Add last_seq column if not exists
	if colCount == 0 {
		alterQuery := fmt.Sprintf("ALTER TABLE %s ADD last_seq VARBINARY(10);", c.checkpointTable)
		_, err = c.dbConn.Exec(alterQuery)
		if err != nil {
			return fmt.Errorf("failed to add last_seq column: %w", err)
		}
		log.Info("Added last_seq column to checkpoints table")
	}

	log.Info("Initialized checkpoints table")
	return nil
}

// LoadLastLSN retrieves the last known LSN for the specified table
func (c *CheckpointManager) LoadLastLSN() ([]byte, []byte, error) {
	var lastLSN, lastSeq []byte

	query := fmt.Sprintf("SELECT last_lsn, last_seq FROM %s WITH (NOLOCK) WHERE table_name = @tableName", c.checkpointTable)
	err := c.dbConn.QueryRow(query, sql.Named("tableName", c.tableName)).Scan(&lastLSN, &lastSeq)
	if err == sql.ErrNoRows {
		startLSNBytes, _ := hex.DecodeString(defaultStartLSN)
		startSeqBytes, _ := hex.DecodeString(defaultStartSEQ)
		lastLSN = startLSNBytes
		lastSeq = startSeqBytes
		log.Info("No previous LSN/SEQ, initializing with defaults", "table", c.tableName, "lsn", hex.EncodeToString(lastLSN), "seq", hex.EncodeToString(lastSeq))
	} else if err != nil {
		return nil, nil, fmt.Errorf("failed to load LSN/SEQ for %s: %w", c.tableName, err)
	} else {
		// If lastSeq is nil, try to get from CDC change table
		if lastSeq == nil {
			cdcTable := fmt.Sprintf("cdc.dbo_%s_CT", c.tableName)
			cdcSeqQuery := fmt.Sprintf(`
				SELECT TOP 1 a.__$seqval 
				FROM %s a 
				JOIN %s b ON a.__$start_lsn = b.last_lsn 
				WHERE b.table_name = @tableName
			`, cdcTable, c.checkpointTable)
			var seqVal []byte
			cdcErr := c.dbConn.QueryRow(cdcSeqQuery, sql.Named("tableName", c.tableName)).Scan(&seqVal)
			if cdcErr == nil && seqVal != nil {
				lastSeq = seqVal
				log.Info("Loaded last_seq from CDC change table", "table", c.tableName, "seq", hex.EncodeToString(lastSeq))
			} else {
				startSeqBytes, _ := hex.DecodeString(defaultStartSEQ)
				lastSeq = startSeqBytes
				log.Info("No last_seq found, initializing with default", "table", c.tableName, "seq", hex.EncodeToString(lastSeq))
			}
		}
	}

	log.Info("Resuming from last LSN", "table", c.tableName, "lsn", hex.EncodeToString(lastLSN), "seq", hex.EncodeToString(lastSeq))
	return lastLSN, lastSeq, nil
}

// SaveLastLSN updates the last known LSN for the specified table
func (c *CheckpointManager) SaveLastLSN(newLSN []byte, lastSeq []byte) error {
	upsertQuery := fmt.Sprintf(`
	MERGE INTO %s AS target
	USING (VALUES (@tableName, @lastLSN, GETDATE(), @lastSeq)) AS source (table_name, last_lsn, updated_at, last_seq)
	ON target.table_name = source.table_name
	WHEN MATCHED THEN 
		UPDATE SET last_lsn = source.last_lsn, updated_at = source.updated_at, last_seq = source.last_seq
	WHEN NOT MATCHED THEN
		INSERT (table_name, last_lsn, updated_at, last_seq) 
		VALUES (source.table_name, source.last_lsn, source.updated_at, source.last_seq);`, c.checkpointTable)

	_, err := c.dbConn.Exec(
		upsertQuery,
		sql.Named("tableName", c.tableName),
		sql.Named("lastLSN", newLSN),
		sql.Named("lastSeq", lastSeq),
	)
	if err != nil {
		return fmt.Errorf("failed to save LSN for %s: %w", c.tableName, err)
	}

	log.Info("Saved new LSN and Seq", "table", c.tableName, "lsn", hex.EncodeToString(newLSN), "seq", hex.EncodeToString(lastSeq))
	return nil
}
