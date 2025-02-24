//go:debug x509negativeserial=1
package testdb

import (
	"database/sql"
	"math/rand"
	"os"
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

func generateRandomPerson() (string, string) {
	return firstNames[rand.Intn(len(firstNames))],
		lastNames[rand.Intn(len(lastNames))]
}

func TestGeneratePersonChanges(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := getTestDB(t)
	defer db.Close()

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Insert 5 new persons
	t.Log("Inserting 5 new persons...")
	for i := 0; i < 5; i++ {
		firstName, lastName := generateRandomPerson()
		_, err := db.Exec(`
			INSERT INTO Persons (FirstName, LastName)
			VALUES (@p1, @p2)`,
			sql.Named("p1", firstName),
			sql.Named("p2", lastName),
		)
		if err != nil {
			t.Fatalf("Failed to insert person: %v", err)
		}
		t.Logf("Inserted: %s %s", firstName, lastName)
		time.Sleep(1 * time.Second) // Wait a bit between inserts
	}

	// Update some existing persons
	t.Log("\nUpdating some persons...")
	rows, err := db.Query("SELECT TOP 3 ID FROM Persons ORDER BY NEWID()")
	if err != nil {
		t.Fatalf("Failed to select random persons: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("Failed to scan ID: %v", err)
		}

		// Generate new random values
		newFirstName, newLastName := generateRandomPerson()
		_, err := db.Exec(`
			UPDATE Persons 
			SET FirstName = @p1, LastName = @p2
			WHERE ID = @p3`,
			sql.Named("p1", newFirstName),
			sql.Named("p2", newLastName),
			sql.Named("p3", id),
		)
		if err != nil {
			t.Fatalf("Failed to update person: %v", err)
		}
		t.Logf("Updated person ID %d: %s %s", id, newFirstName, newLastName)
		time.Sleep(1 * time.Second) // Wait a bit between updates
	}

	// Delete some persons
	t.Log("\nDeleting some persons...")
	rows, err = db.Query("SELECT TOP 2 ID FROM Persons ORDER BY NEWID()")
	if err != nil {
		t.Fatalf("Failed to select random persons: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("Failed to scan ID: %v", err)
		}

		_, err := db.Exec("DELETE FROM Persons WHERE ID = @p1", sql.Named("p1", id))
		if err != nil {
			t.Fatalf("Failed to delete person: %v", err)
		}
		t.Logf("Deleted person ID %d", id)
		time.Sleep(1 * time.Second) // Wait a bit between deletes
	}

	// Final verification
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM Persons").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to get final count: %v", err)
	}
	t.Logf("\nFinal person count in database: %d", count)
}
