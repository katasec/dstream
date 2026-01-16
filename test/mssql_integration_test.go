package test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/katasec/testcontainers-go-presets/mssql"
	_ "github.com/microsoft/go-mssqldb"
	"github.com/testcontainers/testcontainers-go"
)

func Test_MSSQL_Setup(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()

	const pw = "MyTestP@ss123!"

	// Start MSSQL container
	c, err := mssql.Run(ctx, mssql.WithPassword(pw))
	if err != nil {
		t.Fatalf("failed to start MSSQL: %v", err)
	}
	t.Cleanup(func() { c.Terminate(context.Background()) })

	// Connect to master
	db, err := connectToDB(t, ctx, c, pw, "master")
	if err != nil {
		t.Fatalf("failed to connect to master DB: %v", err)
	}
	defer db.Close()

	// Create TestDB
	if _, err := db.ExecContext(ctx, `CREATE DATABASE TestDB`); err != nil {
		t.Fatalf("failed to create TestDB: %v", err)
	}

	// Connect to TestDB using the helper (no string concatenation)
	testDBConnStr, err := mssql.ConnectionString(ctx, c, pw, "TestDB")
	if err != nil {
		t.Fatalf("failed to build TestDB conn string: %v", err)
	}

	testdb, err := sql.Open("sqlserver", testDBConnStr)
	if err != nil {
		t.Fatalf("failed to connect to TestDB: %v", err)
	}
	defer testdb.Close()

	// Create Cars
	createTables(t, testdb, ctx)
	t.Log("Successfully created TestDB, Cars, and Persons tables!")
}

func connectToDB(t *testing.T, ctx context.Context, c testcontainers.Container, pw string, dbName string) (db *sql.DB, err error) {
	connStr, err := mssql.ConnectionString(ctx, c, pw, "master")
	if err != nil {
		t.Fatalf("failed to build master conn string: %v", err)
	}
	db, err = sql.Open("sqlserver", connStr)
	if err != nil {
		t.Fatalf("failed to open sql connection: %v", err)
	}
	return db, err
}

func createTables(t *testing.T, db *sql.DB, ctx context.Context) {

	// Create Cars
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE Cars (
			CarID INT IDENTITY(1,1) NOT NULL PRIMARY KEY,
			BrandName NVARCHAR(50) NOT NULL,
			Color NVARCHAR(30) NOT NULL
		)`); err != nil {
		t.Fatalf("failed to create Cars table: %v", err)
	}

	// Create Persons
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE Persons (
			ID INT IDENTITY(1,1) NOT NULL PRIMARY KEY,
			FirstName VARCHAR(100),
			LastName VARCHAR(100)
		)`); err != nil {
		t.Fatalf("failed to create Persons table: %v", err)
	}
}
