package config

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/katasec/dstream/azureservicebus"
)

// CheckConfig validates the configuration based on the output type and lock type requirements
func (c *Config) CheckConfig() {
	if c.Ingester.DBConnectionString == "" {
		log.Println("Error, DBConnectionString was not found, exiting.")
		os.Exit(0)
	}

	// Validate Output configuration
	switch strings.ToLower(c.Publisher.Output.Type) {
	case "azure_service_bus":
		c.serviceBusConfigCheck()
	case "console":
		// Console output type doesn't need a connection string
		log.Println("Output set to console; no additional connection string required.")
	default:
		log.Fatalf("Error, unknown output type: %s", c.Publisher.Output.Type)
	}

	// Validate Lock configuration
	switch strings.ToLower(c.Ingester.Locks.Type) {
	case "azure_blob_db":
		c.validateBlobLockConfig()
	case "azure_blob":
		c.validateBlobLockConfig()
	default:
		log.Fatalf("Error, unknown lock type: %s", c.Ingester.Locks.Type)
	}
}

// validateBlobLockConfig validates the Azure Blob configuration for locks
func (c *Config) validateBlobLockConfig() {

	// Check for connection string
	connectionString := c.Ingester.Locks.ConnectionString
	if connectionString == "" {
		log.Fatalf("Error, Azure Blob Storage connection string is required for blob locks.")
	}

	// Check for container name
	containerName := c.Ingester.Locks.ContainerName
	if containerName == "" {
		log.Fatalf("Error, Azure Blob Storage container name is required for blob locks.")
	}

	// Create blob client
	client, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		log.Fatalf("Failed to create Azure Blob client: %v", err)
	}

	// Ensure the container exists
	_, err = client.CreateContainer(context.TODO(), containerName, nil)
	if err != nil && !strings.Contains(err.Error(), "ContainerAlreadyExists") {
		log.Fatalf("Failed to ensure Azure Blob container %s: %v", containerName, err)
	}

	log.Printf("Validated Azure Blob container for locks: %s", containerName)
}

// serviceBusConfigCheck validates the Service Bus configuration and ensures topics exist for each table
func (c *Config) serviceBusConfigCheck() {

	connectionString := c.Publisher.Output.ConnectionString
	publisherType := c.Publisher.Output.Type

	if connectionString == "" {
		log.Fatalf("Error, %s connection string is required.", publisherType)
	}

	// Create a Service Bus admin client
	client, err := admin.NewClientFromConnectionString(connectionString, nil)

	if err != nil {
		log.Println(connectionString)
		log.Fatalf("Failed to create Service Bus client: %v", err)
	} else {
		log.Println("Service Bus client created")
	}

	// Ensure each topic exists or create it if not
	for _, table := range c.Ingester.Tables {
		topicName := azureservicebus.GenTopicName(c.Ingester.DBConnectionString, table.Name)
		log.Printf("Ensuring topic exists: %s\n", topicName)

		// Check and create topic if it doesn't exist
		if err := azureservicebus.CreateTopicIfNotExists(client, topicName); err != nil {
			log.Fatalf("Error ensuring topic %s exists: %v", topicName, err)
		}
	}
}
