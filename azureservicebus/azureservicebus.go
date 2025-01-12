package azureservicebus

import (
	"fmt"
	"net/url"
	"os"
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

// ExtractServerName extracts the server name from a connection string
// If the server name is "localhost", it uses the hostname of the current machine.
func extractServerName(connectionString string) (string, error) {
	// Parse the connection string
	u, err := url.Parse(connectionString)
	if err != nil {
		return "", fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Extract the server name from the host part of the URL
	host := u.Host
	if host == "" {
		return "", fmt.Errorf("server name not found in connection string")
	}

	// Split the host to handle cases with a port (e.g., localhost:1433)
	host = strings.Split(host, ":")[0]

	// If the host is "localhost", get the system's hostname
	if strings.ToLower(host) == "localhost" {
		hostname, err := os.Hostname()
		if err != nil {
			return "", fmt.Errorf("failed to get hostname: %w", err)
		}
		host = hostname
	}

	return host, nil
}
