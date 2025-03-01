package servicebus

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
	"github.com/katasec/dstream/internal/utils"
)

// GenTopicName generates a topic name for a given table
func GenTopicName(connectionString string, tableName string) (string, error) {
	dbName, err := extractDatabaseName(connectionString)
	if err != nil {
		return "", fmt.Errorf("failed to extract database name: %w", err)
	}

	serverName, err := utils.ExtractServerNameFromConnectionString(connectionString)
	if err != nil {
		return "", fmt.Errorf("failed to extract server name: %w", err)
	}

	topicName := fmt.Sprintf("%s.%s.%s.events", serverName, dbName, strings.ToLower(tableName))
	return topicName, nil
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

// Deprecated: Use utils.ExtractServerNameFromConnectionString instead
// This function is kept for backward compatibility
func extractServerName(connectionString string) (string, error) {
	return utils.ExtractServerNameFromConnectionString(connectionString)
}

// isIPAddress checks if a string is an IP address or part of one (like '127')
// Deprecated: This function is no longer used directly
func isIPAddress(host string) bool {
	// Check if it's a full IP address
	if ip := net.ParseIP(host); ip != nil {
		return true
	}

	// Check if it's a partial IP (e.g. '127' from '127.0.0.1')
	if _, err := strconv.Atoi(host); err == nil {
		// It's a number, check if it's in valid IP octet range (0-255)
		if num, _ := strconv.Atoi(host); num >= 0 && num <= 255 {
			return true
		}
	}

	return false
}

// CreateTopicIfNotExists checks if a topic exists and creates it if it doesn't
func CreateTopicIfNotExists(client *admin.Client, topicName string) error {
	ctx := context.Background()

	log.Debug("Checking if topic exists", "topic", topicName)

	// Try to get the topic properties
	topicProps, err := client.GetTopic(ctx, topicName, nil)
	if err == nil && topicProps != nil {
		log.Debug("Topic already exists", "topic", topicName)
	} else {
		log.Debug("Creating topic", "topic", topicName)

		// Create the topic since it doesn't exist
		_, err = client.CreateTopic(ctx, topicName, nil)
		if err != nil {
			log.Error("Failed to create topic", "topic", topicName, "error", err)
			return fmt.Errorf("failed to create topic: %w", err)
		}
		log.Debug("Successfully created topic", "topic", topicName)
	}

	// Create subscription if it doesn't exist
	subscriptionName := "sub1"
	log.Debug("Checking if subscription exists", "topic", topicName, "subscription", subscriptionName)

	// Try to get the subscription properties
	subProps, err := client.GetSubscription(ctx, topicName, subscriptionName, nil)
	if err == nil && subProps != nil {
		log.Debug("Subscription already exists", "topic", topicName, "subscription", subscriptionName)
		return nil
	}

	log.Debug("Creating subscription", "topic", topicName, "subscription", subscriptionName)

	// Create the subscription
	_, err = client.CreateSubscription(ctx, topicName, subscriptionName, nil)
	if err != nil {
		log.Error("Failed to create subscription", "topic", topicName, "subscription", subscriptionName, "error", err)
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	log.Debug("Successfully created subscription", "topic", topicName, "subscription", subscriptionName)
	return nil
}
