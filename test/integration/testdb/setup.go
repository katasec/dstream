package testdb

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
)

// TestDB represents a test database instance
type TestDB struct {
	DB *sql.DB
}

// NewTestDB creates a new test database using the connection string from DSTREAM_DB_CONNECTION_STRING
func NewTestDB() (*TestDB, error) {
	// Get connection string from environment
	connStr := os.Getenv("DSTREAM_DB_CONNECTION_STRING")
	if connStr == "" {
		return nil, fmt.Errorf("DSTREAM_DB_CONNECTION_STRING environment variable not set")
	}

	// Connect to master database to create test database
	masterConn, err := sql.Open("sqlserver", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master database: %w", err)
	}
	defer masterConn.Close()

	// Create the test database and enable CDC
	_, err = masterConn.Exec(`
		IF NOT EXISTS (SELECT * FROM sys.databases WHERE name = 'TestDB')
		BEGIN
			CREATE DATABASE TestDB;
		END
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create TestDB: %w", err)
	}

	// Connect to TestDB to enable CDC
	testConnStr := strings.ReplaceAll(connStr, "database=master", "database=TestDB")
	testConnStr = strings.ReplaceAll(testConnStr, "Database=master", "Database=TestDB")
	testConn, err := sql.Open("sqlserver", testConnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to TestDB: %w", err)
	}
	defer testConn.Close()

	// Enable CDC on the database
	_, err = testConn.Exec(`
		IF NOT EXISTS (SELECT 1 FROM sys.databases WHERE name = 'TestDB' AND is_cdc_enabled = 1)
		BEGIN
			EXEC sys.sp_cdc_enable_db;
		END
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to enable CDC on TestDB: %w", err)
	}

	// Create a new connection for the test database
	db, err := sql.Open("sqlserver", testConnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to TestDB: %w", err)
	}

	// Create test tables
	err = createTestTables(db)
	if err != nil {
		testConn.Close()
		return nil, fmt.Errorf("failed to create test tables: %w", err)
	}

	log.Println("Test database and tables created successfully")
	return &TestDB{DB: db}, nil
}

// createTestTables creates the test tables in the database
func createTestTables(db *sql.DB) error {
	// Create Persons table
	_, err := db.Exec(`
		IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'Persons')
		BEGIN
			CREATE TABLE [dbo].[Persons](
				[ID] [int] IDENTITY(1,1) NOT NULL,
				[FirstName] [varchar](100) NULL,
				[LastName] [varchar](100) NULL,
			PRIMARY KEY CLUSTERED 
			(
				[ID] ASC
			)WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, IGNORE_DUP_KEY = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON) ON [PRIMARY]
			) ON [PRIMARY];

			-- Enable CDC on Persons table
			EXEC sys.sp_cdc_enable_table
				@source_schema = 'dbo',
				@source_name = 'Persons',
				@role_name = NULL,
				@supports_net_changes = 0;
		END
	`)
	if err != nil {
		return fmt.Errorf("failed to create Persons table: %w", err)
	}

	// Create Cars table
	_, err = db.Exec(`
		IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'Cars')
		BEGIN
			CREATE TABLE [dbo].[Cars](
				[CarID] [int] IDENTITY(1,1) NOT NULL,
				[BrandName] [nvarchar](50) NOT NULL,
				[Color] [nvarchar](30) NOT NULL,
			PRIMARY KEY CLUSTERED 
			(
				[CarID] ASC
			)WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, IGNORE_DUP_KEY = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON) ON [PRIMARY]
			) ON [PRIMARY];

			-- Enable CDC on Cars table
			EXEC sys.sp_cdc_enable_table
				@source_schema = 'dbo',
				@source_name = 'Cars',
				@role_name = NULL,
				@supports_net_changes = 0;
		END
	`)
	if err != nil {
		return fmt.Errorf("failed to create Cars table: %w", err)
	}

	return nil
}

// Close closes the database connection
func (tdb *TestDB) Close() error {
	return tdb.DB.Close()
}

// Reset cleans up the test data but keeps the tables
func (tdb *TestDB) Reset() error {
	_, err := tdb.DB.Exec(`
		DELETE FROM dbo.Persons;
		DELETE FROM dbo.Cars;
	`)
	return err
}

// InsertTestData inserts some test data into the tables
func (tdb *TestDB) InsertTestData() error {
	// Insert test persons
	_, err := tdb.DB.Exec(`
		INSERT INTO dbo.Persons (FirstName, LastName)
		VALUES 
			('John', 'Doe'),
			('Jane', 'Smith'),
			('Bob', 'Johnson');
	`)
	if err != nil {
		return fmt.Errorf("failed to insert test persons: %w", err)
	}

	// Insert test cars
	_, err = tdb.DB.Exec(`
		INSERT INTO dbo.Cars (BrandName, Color)
		VALUES 
			('Toyota', 'Red'),
			('Honda', 'Blue'),
			('Ford', 'Black');
	`)
	if err != nil {
		return fmt.Errorf("failed to insert test cars: %w", err)
	}

	return nil
}

// Drop drops the test database
func (tdb *TestDB) Drop() error {
	// First close our connection to the test database
	err := tdb.Close()
	if err != nil {
		return fmt.Errorf("failed to close test database connection: %w", err)
	}

	// Connect to master to drop the test database
	masterConn, err := sql.Open("sqlserver", os.Getenv("DSTREAM_DB_CONNECTION_STRING"))
	if err != nil {
		return fmt.Errorf("failed to connect to master database: %w", err)
	}
	defer masterConn.Close()

	// Drop the test database
	_, err = masterConn.Exec(`
		IF EXISTS (SELECT * FROM sys.databases WHERE name = 'TestDB')
		BEGIN
			ALTER DATABASE TestDB SET SINGLE_USER WITH ROLLBACK IMMEDIATE;
			DROP DATABASE TestDB;
		END
	`)
	if err != nil {
		return fmt.Errorf("failed to drop TestDB: %w", err)
	}

	log.Println("Test database dropped successfully")
	return nil
}
