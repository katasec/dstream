package testdb

import (
	"os"
	"testing"

	_ "github.com/microsoft/go-mssqldb"
)

func TestDatabaseSetup(t *testing.T) {
	// Skip if connection string is not set
	if os.Getenv("DSTREAM_DB_CONNECTION_STRING") == "" {
		t.Skip("DSTREAM_DB_CONNECTION_STRING not set")
	}

	// Create test database
	testDB, err := NewTestDB()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	// Keeping the database for repeated runs
	// defer testDB.Drop() // Clean up after test

	// Insert test data
	err = testDB.InsertTestData()
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Verify Persons table has data
	var personCount int
	err = testDB.DB.QueryRow("SELECT COUNT(*) FROM dbo.Persons").Scan(&personCount)
	if err != nil {
		t.Fatalf("Failed to count persons: %v", err)
	}
	if personCount != 3 {
		t.Errorf("Expected 3 persons, got %d", personCount)
	}

	// Verify Cars table has data
	var carCount int
	err = testDB.DB.QueryRow("SELECT COUNT(*) FROM dbo.Cars").Scan(&carCount)
	if err != nil {
		t.Fatalf("Failed to count cars: %v", err)
	}
	if carCount != 3 {
		t.Errorf("Expected 3 cars, got %d", carCount)
	}

	// Verify CDC is enabled on database
	var isCDCEnabled bool
	err = testDB.DB.QueryRow("SELECT is_cdc_enabled FROM sys.databases WHERE name = 'TestDB'").Scan(&isCDCEnabled)
	if err != nil {
		t.Fatalf("Failed to check CDC status: %v", err)
	}
	if !isCDCEnabled {
		t.Error("CDC is not enabled on the database")
	}

	// Verify CDC tables exist
	var carsTableExists, personsTableExists bool
	err = testDB.DB.QueryRow("SELECT CASE WHEN COUNT(*) > 0 THEN 1 ELSE 0 END FROM cdc.change_tables WHERE source_object_id = OBJECT_ID('dbo.Cars')").Scan(&carsTableExists)
	if err != nil {
		t.Fatalf("Failed to check Cars CDC table: %v", err)
	}
	if !carsTableExists {
		t.Error("CDC table for Cars does not exist")
	}

	err = testDB.DB.QueryRow("SELECT CASE WHEN COUNT(*) > 0 THEN 1 ELSE 0 END FROM cdc.change_tables WHERE source_object_id = OBJECT_ID('dbo.Persons')").Scan(&personsTableExists)
	if err != nil {
		t.Fatalf("Failed to check Persons CDC table: %v", err)
	}
	if !personsTableExists {
		t.Error("CDC table for Persons does not exist")
	}

	// Test Reset functionality
	err = testDB.Reset()
	if err != nil {
		t.Fatalf("Failed to reset test data: %v", err)
	}

	// Verify tables are empty
	err = testDB.DB.QueryRow("SELECT COUNT(*) FROM dbo.Persons").Scan(&personCount)
	if err != nil {
		t.Fatalf("Failed to count persons after reset: %v", err)
	}
	if personCount != 0 {
		t.Errorf("Expected 0 persons after reset, got %d", personCount)
	}

	err = testDB.DB.QueryRow("SELECT COUNT(*) FROM dbo.Cars").Scan(&carCount)
	if err != nil {
		t.Fatalf("Failed to count cars after reset: %v", err)
	}
	if carCount != 0 {
		t.Errorf("Expected 0 cars after reset, got %d", carCount)
	}
}
