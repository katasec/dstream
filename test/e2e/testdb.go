package e2e

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/katasec/testcontainers-go-presets/mssql"
	_ "github.com/microsoft/go-mssqldb"
	"github.com/testcontainers/testcontainers-go"
)

const testPassword = "MyTestP@ss123!"

// TestDB manages a containerized SQL Server with CDC enabled.
type TestDB struct {
	Container testcontainers.Container
	DB        *sql.DB
	ConnStr   string // connection string to TestDB (not master)
	Host      string // container host
	Port      string // mapped port
}

// StartTestDB spins up a SQL Server container, creates TestDB with CDC,
// and creates the Cars and Persons tables with CDC tracking enabled.
func StartTestDB(t *testing.T, ctx context.Context) *TestDB {
	t.Helper()

	// Start MSSQL container
	c, err := mssql.Run(ctx, mssql.WithPassword(testPassword))
	if err != nil {
		t.Fatalf("failed to start MSSQL container: %v", err)
	}
	t.Cleanup(func() { c.Terminate(context.Background()) })

	// Get container host and port for connection string building
	host, err := c.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}
	mappedPort, err := c.MappedPort(ctx, "1433/tcp")
	if err != nil {
		t.Fatalf("failed to get mapped port: %v", err)
	}

	// Connect to master to create TestDB
	masterConnStr, err := mssql.ConnectionString(ctx, c, testPassword, "master")
	if err != nil {
		t.Fatalf("failed to build master conn string: %v", err)
	}

	masterDB, err := sql.Open("sqlserver", masterConnStr)
	if err != nil {
		t.Fatalf("failed to connect to master: %v", err)
	}
	defer masterDB.Close()

	if _, err := masterDB.ExecContext(ctx, `CREATE DATABASE TestDB`); err != nil {
		t.Fatalf("failed to create TestDB: %v", err)
	}

	// Connect to TestDB
	testConnStr, err := mssql.ConnectionString(ctx, c, testPassword, "TestDB")
	if err != nil {
		t.Fatalf("failed to build TestDB conn string: %v", err)
	}

	db, err := sql.Open("sqlserver", testConnStr)
	if err != nil {
		t.Fatalf("failed to connect to TestDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Enable CDC on database
	if _, err := db.ExecContext(ctx, `EXEC sys.sp_cdc_enable_db`); err != nil {
		t.Fatalf("failed to enable CDC on TestDB: %v", err)
	}

	// Create tables and enable CDC on each
	createTablesWithCDC(t, ctx, db)

	tdb := &TestDB{
		Container: c,
		DB:        db,
		ConnStr:   testConnStr,
		Host:      host,
		Port:      mappedPort.Port(),
	}

	t.Logf("TestDB ready: %s:%s (CDC enabled, 2 tables)", host, mappedPort.Port())
	return tdb
}

func createTablesWithCDC(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()

	tables := []struct {
		name   string
		create string
	}{
		{
			name: "Persons",
			create: `CREATE TABLE dbo.Persons (
				ID INT IDENTITY(1,1) NOT NULL PRIMARY KEY,
				FirstName VARCHAR(100),
				LastName VARCHAR(100)
			)`,
		},
		{
			name: "Cars",
			create: `CREATE TABLE dbo.Cars (
				CarID INT IDENTITY(1,1) NOT NULL PRIMARY KEY,
				BrandName NVARCHAR(50) NOT NULL,
				Color NVARCHAR(30) NOT NULL
			)`,
		},
	}

	for _, table := range tables {
		if _, err := db.ExecContext(ctx, table.create); err != nil {
			t.Fatalf("failed to create %s table: %v", table.name, err)
		}

		cdcSQL := fmt.Sprintf(`EXEC sys.sp_cdc_enable_table
			@source_schema = 'dbo',
			@source_name = '%s',
			@role_name = NULL,
			@supports_net_changes = 0`, table.name)

		if _, err := db.ExecContext(ctx, cdcSQL); err != nil {
			t.Fatalf("failed to enable CDC on %s: %v", table.name, err)
		}
	}

	t.Log("Tables created with CDC: Persons, Cars")
}

// InsertTestData inserts sample rows to generate CDC change events.
func (tdb *TestDB) InsertTestData(t *testing.T, ctx context.Context) {
	t.Helper()

	_, err := tdb.DB.ExecContext(ctx, `
		INSERT INTO dbo.Persons (FirstName, LastName) VALUES
			('John', 'Doe'),
			('Jane', 'Smith'),
			('Bob', 'Johnson')
	`)
	if err != nil {
		t.Fatalf("failed to insert persons: %v", err)
	}

	_, err = tdb.DB.ExecContext(ctx, `
		INSERT INTO dbo.Cars (BrandName, Color) VALUES
			('Toyota', 'Red'),
			('Honda', 'Blue'),
			('Ford', 'Black')
	`)
	if err != nil {
		t.Fatalf("failed to insert cars: %v", err)
	}

	t.Log("Inserted 3 persons and 3 cars")
}

// DStreamConnectionString returns a connection string in the format the MSSQL
// ingester expects (semicolon-delimited, no URL scheme).
func (tdb *TestDB) DStreamConnectionString() string {
	return fmt.Sprintf(
		"server=%s,%s;user id=sa;password=%s;database=TestDB;encrypt=disable",
		tdb.Host, tdb.Port, testPassword,
	)
}
