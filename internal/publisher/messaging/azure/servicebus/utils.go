package servicebus

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
)

// GenTopicName generates a topic name for a given table
func GenTopicName(connectionString string, tableName string) string {
	dbName, err := extractDatabaseName(connectionString)
	if err != nil {
		log.Error("Database name not found in connection string")
		os.Exit(1)
	}

	serverName, err := extractServerName(connectionString)
	if err != nil {
		log.Error("Server name not found in connection string")
		os.Exit(1)
	}

	topicName := fmt.Sprintf("%s.%s.%s.events", serverName, dbName, strings.ToLower(tableName))
	return topicName
}

// extractDatabaseName extracts the database name from a connection string
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

// extractServerName extracts the server name from a connection string
func extractServerName(connectionString string) (string, error) {
	// Parse the connection string
	u, err := url.Parse(connectionString)
	if err != nil {
		return "", fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Get the server name from the host
	serverName := strings.Split(u.Host, ".")[0]
	if serverName == "" {
		return "", fmt.Errorf("server name not found in connection string")
	}

	// If the server is localhost, use the machine's hostname
	if strings.ToLower(serverName) == "localhost" {
		hostname, err := os.Hostname()
		if err != nil {
			return "", fmt.Errorf("failed to get hostname: %w", err)
		}
		serverName = hostname
	}

	return strings.ToLower(serverName), nil
}

// CreateTopicIfNotExists checks if a topic exists and creates it if it doesn't
func CreateTopicIfNotExists(client *admin.Client, topicName string) error {
	ctx := context.Background()

	// Try to get the topic properties
	_, err := client.GetTopic(ctx, topicName, nil)
	if err == nil {
		// Topic exists
		return nil
	}

	// Create the topic if it doesn't exist
	_, err = client.CreateTopic(ctx, topicName, nil)
	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	return nil
}
