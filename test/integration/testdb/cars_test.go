package testdb

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/microsoft/go-mssqldb"
)

// Available car brands and colors for random selection
var (
	carBrands = []string{
		"Toyota", "Honda", "Ford", "BMW", "Mercedes",
		"Audi", "Volkswagen", "Tesla", "Porsche", "Lexus",
		"Hyundai", "Kia", "Mazda", "Subaru", "Chevrolet",
	}

	carColors = []string{
		"Red", "Blue", "Black", "White", "Silver",
		"Gray", "Green", "Yellow", "Orange", "Purple",
		"Brown", "Gold", "Bronze", "Pearl", "Navy",
	}
)

func TestRandomCarInsertion(t *testing.T) {
	// Skip if connection string is not set
	if os.Getenv("DSTREAM_DB_CONNECTION_STRING") == "" {
		t.Skip("DSTREAM_DB_CONNECTION_STRING not set")
	}

	// Create test database
	testDB, err := NewTestDB()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Initialize random seed with current time
	rand.Seed(time.Now().UnixNano())

	// Create a slice to store our random car data
	type carData struct {
		brand string
		color string
	}
	cars := make([]carData, 5)

	// Generate 5 random unique combinations
	usedCombos := make(map[string]bool)
	for i := 0; i < 5; i++ {
		var car carData
		var combo string
		for {
			car = carData{
				brand: carBrands[rand.Intn(len(carBrands))],
				color: carColors[rand.Intn(len(carColors))],
			}
			combo = fmt.Sprintf("%s-%s", car.brand, car.color)
			if !usedCombos[combo] {
				usedCombos[combo] = true
				break
			}
		}
		cars[i] = car
	}

	// Build and execute the insert query with parameterized values
	values := []interface{}{}
	paramPlaceholders := []string{}
	for i, car := range cars {
		values = append(values, car.brand, car.color)
		paramPlaceholders = append(paramPlaceholders, fmt.Sprintf("(@p%d, @p%d)", i*2+1, i*2+2))
	}

	query := fmt.Sprintf(
		"INSERT INTO dbo.Cars (BrandName, Color) VALUES %s",
		strings.Join(paramPlaceholders, ","),
	)

	_, err = testDB.DB.Exec(query, values...)
	if err != nil {
		t.Fatalf("Failed to insert random cars: %v", err)
	}

	// Verify the insertions
	rows, err := testDB.DB.Query("SELECT BrandName, Color FROM dbo.Cars WHERE CarID > (SELECT MAX(CarID) - 5 FROM dbo.Cars)")
	if err != nil {
		t.Fatalf("Failed to query inserted cars: %v", err)
	}
	defer rows.Close()

	// Print the inserted cars
	t.Log("Inserted the following random cars:")
	for rows.Next() {
		var brand, color string
		if err := rows.Scan(&brand, &color); err != nil {
			t.Fatalf("Failed to scan row: %v", err)
		}
		t.Logf("- %s (%s)", brand, color)
	}
}
