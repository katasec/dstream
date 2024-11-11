package azureservicebus

import (
	"fmt"
	"net/url"
	"strings"
)

func GenTopicName(connectionString string, tableName string) string {
	dbName, _ := extractDatabaseName(connectionString)
	return fmt.Sprintf("%s-%s-events", dbName, strings.ToLower(tableName))
}

// ExtractDatabaseName extracts the database name from a connection string
func extractDatabaseName(connectionString string) (string, error) {
	// Parse the connection string
	u, err := url.Parse(connectionString)
	if err != nil {
		return "", fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Look for the "database" query parameter
	dbName := u.Query().Get("database")
	if dbName == "" {
		return "", fmt.Errorf("database name not found in connection string")
	}
	dbName = strings.ToLower(dbName)
	return dbName, nil
}
