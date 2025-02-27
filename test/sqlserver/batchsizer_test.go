//go:debug x509negativeserial=1

package sqlserver_test

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/katasec/dstream/internal/cdc/sqlserver"
	_ "github.com/microsoft/go-mssqldb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchSizer(t *testing.T) {
	// Setup test database connection
	connStr := os.Getenv("DSTREAM_DB_CONNECTION_STRING")
	require.NotEmpty(t, connStr, "DSTREAM_DB_CONNECTION_STRING environment variable must be set")

	db, err := sql.Open("sqlserver", connStr)
	require.NoError(t, err)
	defer db.Close()

	// Create test table and enable CDC
	setupTestBatchSizeTable(t, db)
	defer cleanupTestBatchSizeTable(t, db)

	tests := []struct {
		name          string
		sku           int   // SKU limit to test
		expectedBatch int32 // Expected batch size for this SKU
	}{
		{
			name:          "Standard SKU Batch Size",
			sku:           sqlserver.StandardSKULimit,
			expectedBatch: 100,
		},
		{
			name:          "Premium SKU Batch Size",
			sku:           sqlserver.PremiumSKULimit,
			expectedBatch: 250,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create BatchSizer for this SKU
			sizer := sqlserver.NewBatchSizer(db, "TestBatchSizeTable", tt.sku)

			// Insert some test records to ensure CDC is working
			insertTestRecords(t, db, 10, 1024) // Just need a few records

			// Wait for CDC to capture changes
			var cdcCount int
			for i := 0; i < 10; i++ {
				err := db.QueryRow(`
					SELECT COUNT(*)
					FROM cdc.dbo_TestBatchSizeTable_CT
				`).Scan(&cdcCount)
				require.NoError(t, err)
				if cdcCount > 0 {
					break
				}
				time.Sleep(time.Second)
			}
			assert.Greater(t, cdcCount, 0, "CDC should have captured some records")

			// Start the sizer
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := sizer.Start(ctx)
			require.NoError(t, err)

			// Wait for initial sampling
			time.Sleep(time.Second)

			// Get batch size and verify it matches the SKU's expected batch size
			batchSize := sizer.GetBatchSize()
			assert.Equal(t, tt.expectedBatch, batchSize, "Batch size should match SKU limit")

			// Verify metrics
			metrics := sizer.GetMetrics()
			assert.Equal(t, batchSize, metrics.CurrentBatchSize)
			assert.NotZero(t, metrics.LastSampleTime)
			assert.NotZero(t, metrics.LastSampleSize)
			assert.NotZero(t, metrics.AvgRowSize)
		})
	}
}

func setupTestBatchSizeTable(t *testing.T, db *sql.DB) {
	// Drop table if it exists
	_, err := db.Exec(`
		IF OBJECT_ID('dbo.TestBatchSizeTable', 'U') IS NOT NULL
		BEGIN
			-- Disable CDC first if it's enabled
			IF EXISTS (SELECT 1 FROM sys.tables t WHERE t.name = 'TestBatchSizeTable' AND t.is_tracked_by_cdc = 1)
			BEGIN
				EXEC sys.sp_cdc_disable_table
					@source_schema = N'dbo',
					@source_name = N'TestBatchSizeTable',
					@capture_instance = 'all'
			END
			-- Drop the table
			DROP TABLE dbo.TestBatchSizeTable
		END
	`)
	require.NoError(t, err)

	// Create test table
	_, err = db.Exec(`
		CREATE TABLE dbo.TestBatchSizeTable (
			ID INT PRIMARY KEY,
			Data NVARCHAR(MAX)
		)
	`)
	require.NoError(t, err)

	// Verify table was created
	var tableExists bool
	err = db.QueryRow(`
		SELECT 1 FROM sys.tables WHERE name = 'TestBatchSizeTable' AND schema_id = SCHEMA_ID('dbo')
	`).Scan(&tableExists)
	if err == sql.ErrNoRows || !tableExists {
		t.Fatal("Failed to create TestBatchSizeTable")
	}
	require.NoError(t, err)

	// Enable CDC on database if not already enabled
	_, err = db.Exec(`
		IF NOT EXISTS (SELECT 1 FROM sys.databases WHERE database_id = DB_ID() AND is_cdc_enabled = 1)
		BEGIN
			EXEC sys.sp_cdc_enable_db
		END
	`)
	require.NoError(t, err)

	// Enable CDC on table
	_, err = db.Exec(`
		EXEC sys.sp_cdc_enable_table
			@source_schema = N'dbo',
			@source_name = N'TestBatchSizeTable',
			@role_name = NULL
	`)
	require.NoError(t, err)

	// Verify CDC is enabled on the table
	var isEnabled bool
	err = db.QueryRow(`
		SELECT 1
		FROM sys.tables t
		JOIN sys.schemas s ON t.schema_id = s.schema_id
		WHERE s.name = 'dbo'
		AND t.name = 'TestBatchSizeTable'
		AND t.is_tracked_by_cdc = 1
	`).Scan(&isEnabled)

	if err == sql.ErrNoRows || !isEnabled {
		t.Fatal("CDC was not properly enabled on TestBatchSizeTable")
	}
	require.NoError(t, err)
	require.NoError(t, err)
}

func cleanupTestBatchSizeTable(t *testing.T, db *sql.DB) {
	// Disable CDC
	_, err := db.Exec(`
		IF EXISTS (SELECT * FROM sys.tables WHERE name = 'TestBatchSizeTable_CT')
		BEGIN
			EXEC sys.sp_cdc_disable_table
			@source_schema = 'dbo',
			@source_name = 'TestBatchSizeTable'
		END
	`)
	require.NoError(t, err)

	// Drop test table
	_, err = db.Exec(`
		IF EXISTS (SELECT * FROM sys.tables WHERE name = 'TestBatchSizeTable')
		DROP TABLE TestBatchSizeTable
	`)
	require.NoError(t, err)
}

func TestCreateTable(t *testing.T) {
	// Get connection string from environment
	connStr := os.Getenv("DSTREAM_DB_CONNECTION_STRING")
	require.NotEmpty(t, connStr, "DSTREAM_DB_CONNECTION_STRING environment variable must be set")

	db, err := sql.Open("sqlserver", connStr)
	require.NoError(t, err)
	defer db.Close()

	// Drop table if it exists
	_, err = db.Exec(`
		IF OBJECT_ID('dbo.TestBatchSizeTable', 'U') IS NOT NULL
		BEGIN
			DROP TABLE dbo.TestBatchSizeTable
		END
	`)
	require.NoError(t, err)

	// Create test table
	_, err = db.Exec(`
		CREATE TABLE dbo.TestBatchSizeTable (
			ID INT PRIMARY KEY,
			Data NVARCHAR(MAX)
		)
	`)
	require.NoError(t, err)

	// Verify table exists
	var tableName string
	err = db.QueryRow(`
		SELECT t.name
		FROM sys.tables t
		JOIN sys.schemas s ON t.schema_id = s.schema_id
		WHERE s.name = 'dbo'
		AND t.name = 'TestBatchSizeTable'
	`).Scan(&tableName)

	require.NoError(t, err)
	assert.Equal(t, "TestBatchSizeTable", tableName)

	// Cleanup
	_, err = db.Exec(`DROP TABLE dbo.TestBatchSizeTable`)
	require.NoError(t, err)
}

func insertTestRecords(t *testing.T, db *sql.DB, count int, size int) {
	// Clear existing data
	_, err := db.Exec("DELETE FROM dbo.TestBatchSizeTable")
	require.NoError(t, err)

	// Generate a string of specified size
	data := strings.Repeat("A", size)

	// Insert records
	for i := 0; i < count; i++ {
		_, err := db.Exec(
			"INSERT INTO dbo.TestBatchSizeTable (ID, Data) VALUES (@p1, @p2)",
			sql.Named("p1", i),
			sql.Named("p2", data),
		)
		require.NoError(t, err)
	}

	// Verify record count
	var actualCount int
	err = db.QueryRow("SELECT COUNT(*) FROM dbo.TestBatchSizeTable").Scan(&actualCount)
	require.NoError(t, err)
	assert.Equal(t, count, actualCount, "Number of inserted records does not match")
}
