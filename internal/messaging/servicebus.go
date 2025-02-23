package messaging

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
	"github.com/katasec/dstream/internal/logging"
)

var log = logging.GetLogger()

func GenTopicName(connectionString string, tableName string) string {
	dbName, err := extractDatabaseName(connectionString)
	if err != nil {
		fmt.Println("The connection string was:" + connectionString)
		log.Error("Database name not found in connection string")
		os.Exit(1)
	}

	serverName, err := extractServerName(connectionString)
	if err != nil {
		fmt.Println("The connection string was:" + connectionString)
		log.Error("Server name not found in connection string")
		os.Exit(1)
	}

	topicName := fmt.Sprintf("%s.%s.%s.events", serverName, dbName, strings.ToLower(tableName))
	return topicName
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

	host = strings.ToLower(host)
	return host, nil
}

// createTopicIfNotExists checks if a topic exists and creates it if it doesn’t
func CreateTopicIfNotExists(client *admin.Client, topicName string) error {

	// If topic does not exist, create it
	log.Info("Creating topic", "topic", topicName)
	response, err := client.CreateTopic(context.TODO(), topicName, nil)

	// Check alreadyExists error
	alreadyExists := false
	if err != nil {
		alreadyExists = strings.Contains(err.Error(), "409 Conflict")
	}

	if alreadyExists {
		log.Debug("Topic already exists", "topic", topicName)
		return nil
	} else if err != nil {
		log.Error("Failed to create topic", "topic", topicName, "error", err)
		// return fmt.Errorf("failed to create topic %s: %w", topicName, err)
		os.Exit(1)
	}
	fmt.Printf("Topic %s created successfully. Status: %d\n", topicName, response.Status)
	return nil
}
