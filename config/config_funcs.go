package config

import (
	"context"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/katasec/dstream/azureservicebus"
)

// CheckConfig validates the configuration based on the output type and lock type requirements
func (c *Config) CheckConfig() {
	if c.Ingester.DBConnectionString == "" {
		log.Error("DBConnectionString not found, exiting")
		os.Exit(0)
	}

	// Validate Output configuration
	switch strings.ToLower(c.Publisher.Output.Type) {
	case "azure_service_bus":
		c.serviceBusConfigCheck()
	case "console":
		// Console output type doesn't need a connection string
		log.Debug("Output set to console; no additional connection string required")
	default:
		log.Error("Unknown output type", "type", c.Publisher.Output.Type)
		os.Exit(1)
	}

	// Validate Lock configuration
	switch strings.ToLower(c.Ingester.Locks.Type) {
	case "azure_blob_db":
		c.validateBlobLockConfig()
	case "azure_blob":
		c.validateBlobLockConfig()
	default:
		log.Error("Unknown lock type", "type", c.Ingester.Locks.Type)
		os.Exit(1)
	}

	// Validate Ingestion connection string
	if c.Ingester.Queue.ConnectionString == "" {
		log.Error("Ingester queue connection string required for ingestion")
		os.Exit(1)
	}

}

// validateBlobLockConfig validates the Azure Blob configuration for locks
func (c *Config) validateBlobLockConfig() {

	// Check for connection string
	connectionString := c.Ingester.Locks.ConnectionString
	if connectionString == "" {
		log.Error("Azure Blob Storage connection string required for blob locks")
		os.Exit(1)
	}

	// Check for container name
	containerName := c.Ingester.Locks.ContainerName
	if containerName == "" {
		log.Error("Azure Blob Storage container name required for blob locks")
		os.Exit(1)
	}

	// Create blob client
	client, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		log.Error("Failed to create Azure Blob client", "error", err)
		os.Exit(1)
	}

	// Ensure the container exists
	_, err = client.CreateContainer(context.TODO(), containerName, nil)
	if err != nil && !strings.Contains(err.Error(), "ContainerAlreadyExists") {
		log.Error("Failed to ensure Azure Blob container", "container", containerName, "error", err)
		os.Exit(1)
	}

	log.Info("Validated Azure Blob container for locks", "container", containerName)
}

// serviceBusConfigCheck validates the Service Bus configuration and ensures topics exist for each table
func (c *Config) serviceBusConfigCheck() {

	connectionString := c.Publisher.Output.ConnectionString
	publisherType := c.Publisher.Output.Type

	if connectionString == "" {
		log.Error("Connection string required", "type", publisherType)
		os.Exit(1)
	}

	// Create a Service Bus admin client
	client, err := admin.NewClientFromConnectionString(connectionString, nil)

	if err != nil {
		log.Debug("Using connection string", "connectionString", connectionString)
		log.Error("Failed to create Service Bus client", "error", err)
		os.Exit(1)
	} else {
		log.Debug("Service Bus client created")
	}

	// Ensure each topic exists or create it if not
	for _, table := range c.Ingester.Tables {
		topicName := azureservicebus.GenTopicName(c.Ingester.DBConnectionString, table.Name)
		log.Info("Ensuring topic exists", "topic", topicName)

		// Check and create topic if it doesn't exist
		if err := azureservicebus.CreateTopicIfNotExists(client, topicName); err != nil {
			log.Error("Failed to ensure topic exists", "topic", topicName, "error", err)
			os.Exit(1)
		}
	}

	// Create a Service Bus admin client for ingester queue
	ingestClient, err := admin.NewClientFromConnectionString(c.Ingester.Queue.ConnectionString, nil)
	if err != nil {
		log.Debug("Using connection string", "connectionString", c.Ingester.Queue.ConnectionString)
		log.Error("Failed to create Service Bus client for ingester", "error", err)
		os.Exit(1)
	}

	// Create the ingest queue if it doesn't exist
	_, err = ingestClient.CreateQueue(context.TODO(), c.Ingester.Queue.Name, nil)
	if err != nil && !strings.Contains(err.Error(), "QueueAlreadyExists") {
		log.Error("Failed to create ingest queue", "queue", c.Ingester.Queue.Name, "error", err)
		os.Exit(1)
	}
	log.Info("Validated ingest queue", "queue", c.Ingester.Queue.Name)
}
