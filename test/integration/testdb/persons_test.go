//go:debug x509negativeserial=1
package testdb

import (
	"database/sql"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/microsoft/go-mssqldb"
)

var (
	firstNames = []string{"John", "Jane", "Bob", "Alice", "Charlie", "Diana", "Edward", "Fiona"}
	lastNames  = []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis"}
)

func getTestDB(t *testing.T) *sql.DB {
	connStr := os.Getenv("DSTREAM_DB_CONNECTION_STRING")
	if connStr == "" {
		t.Fatal("DSTREAM_DB_CONNECTION_STRING environment variable not set")
	}

	db, err := sql.Open("sqlserver", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	return db
}

func generateRandomPerson(rng *rand.Rand) (string, string) {
	return firstNames[rng.Intn(len(firstNames))],
		lastNames[rng.Intn(len(lastNames))]
}

func TestSimplePersonChanges(t *testing.T) {
	// Configure the number of persons to insert
	numPersonsToInsert := 10 // Change this value to 1, 10, or 100 as needed

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get database connection
	db := getTestDB(t)
	defer db.Close()

	// Check that we're connected to TestDB
	var dbName string
	err := db.QueryRow("SELECT DB_NAME()").Scan(&dbName)
	if err != nil {
		t.Fatalf("Failed to get database name: %v", err)
	}
	if strings.ToLower(dbName) != strings.ToLower("TestDB") {
		t.Fatalf("Expected to be connected to TestDB, but connected to %s", dbName)
	}
	t.Logf("Connected to database: %s", dbName)

	// Reset the database to ensure we have a clean state
	testDB, err := NewTestDB()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer testDB.Close()

	// Reset and insert initial test data
	t.Log("Resetting database...")
	err = testDB.Reset()
	if err != nil {
		t.Fatalf("Failed to reset database: %v", err)
	}

	// Insert initial test data
	t.Log("Inserting initial test data...")
	err = testDB.InsertTestData()
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Check if the table exists and is empty
	var tableCount int
	err = testDB.DB.QueryRow("SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'Persons'").Scan(&tableCount)
	if err != nil {
		t.Fatalf("Failed to check if Persons table exists: %v", err)
	}
	if tableCount == 0 {
		t.Fatalf("Persons table does not exist")
	}

	// Check initial count
	var initialCount int
	err = testDB.DB.QueryRow("SELECT COUNT(*) FROM Persons").Scan(&initialCount)
	if err != nil {
		t.Fatalf("Failed to get initial count: %v", err)
	}
	t.Logf("Initial person count: %d", initialCount)

	// Initialize random generator
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Insert multiple persons based on numPersonsToInsert
	t.Logf("Inserting %d new persons...", numPersonsToInsert)

	for i := 0; i < numPersonsToInsert; i++ {
		firstName, lastName := generateRandomPerson(rng)
		result, err := testDB.DB.Exec(`
			INSERT INTO Persons (FirstName, LastName)
			VALUES (@p1, @p2)`,
			sql.Named("p1", firstName),
			sql.Named("p2", lastName),
		)
		if err != nil {
			t.Fatalf("Failed to insert person: %v", err)
		}

		// Check if the insert was successful
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			t.Fatalf("Failed to get rows affected: %v", err)
		}
		if rowsAffected != 1 {
			t.Fatalf("Expected 1 row to be affected, got %d", rowsAffected)
		}

		t.Logf("Inserted: %s %s", firstName, lastName)
	}

	// Verify final count
	var finalCount int
	err = testDB.DB.QueryRow("SELECT COUNT(*) FROM Persons").Scan(&finalCount)
	if err != nil {
		t.Fatalf("Failed to get final count: %v", err)
	}
	t.Logf("Final person count: %d (should be %d more than initial)", finalCount, numPersonsToInsert)

	if finalCount != initialCount+numPersonsToInsert {
		t.Errorf("Expected %d persons, got %d", initialCount+numPersonsToInsert, finalCount)
	}
}
