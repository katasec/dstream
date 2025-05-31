package config

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Masterminds/sprig"
	_ "github.com/denisenkom/go-mssqldb" // SQL Server driver
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/katasec/dstream/internal/publisher/messaging/azure/servicebus"
	"github.com/katasec/dstream/pkg/logging"
)

// CheckConfig validates the configuration based on the output type and lock type requirements
func (c *Config) CheckConfig() {
	if c.Ingester.DBConnectionString == "" {
		log.Error("DBConnectionString not found, exiting")
		os.Exit(0)
	}

	// Validate Output configuration
	switch strings.ToLower(c.Router.Output.Type) {
	case "azure_service_bus":
		c.serviceBusConfigCheck()
	case "console":
		// Console output type doesn't need a connection string
		log.Debug("Output set to console; no additional connection string required")
	default:
		log.Error("Unknown output type", "type", c.Router.Output.Type)
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

	isEnabled, err := c.checkDatabaseCDCEnabled()
	if err != nil {
		log.Error("Failed to check CDC status", "error", err)
		os.Exit(1)
	}
	if !isEnabled {
		log.Error("CDC is not enabled for the database")
		os.Exit(1)
	}
}

// checkDatabaseCDCEnabled queries the database to verify if CDC is enabled at the database level.
// It logs the status and returns true if enabled, false otherwise, along with any error.
func (c *Config) checkDatabaseCDCEnabled() (bool, error) {
	// Check if CDC is enabled at the database level
	db, err := sql.Open("sqlserver", c.Ingester.DBConnectionString)
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Test the connection
	err = db.Ping()
	if err != nil {
		logging.GetLogger().Error("Failed to ping database", "error", err)
	} else {
		logging.GetLogger().Debug("Successfully pinged database")
	}

	query := `SELECT is_cdc_enabled FROM sys.databases WHERE name = DB_NAME();`
	var isEnabled bool

	// Use a short context for this check
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = db.QueryRowContext(ctx, query).Scan(&isEnabled)
	if err != nil {
		log.Error("Failed to query database for CDC status", "error", err)
		return false, fmt.Errorf("error checking database CDC status: %w", err)
	}

	if !isEnabled {
		log.Error("CDC is NOT enabled for the database specified in DBConnectionString.")
		return false, fmt.Errorf("CDC is not enabled for the database")
	}

	log.Info("CDC is enabled for the database.")
	return true, nil
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

	connectionString := c.Router.Output.ConnectionString
	publisherType := c.Router.Output.Type

	if connectionString == "" {
		log.Error("Connection string required", "type", publisherType)
		os.Exit(1)
	}

	// Create a Service Bus admin client
	client, err := admin.NewClientFromConnectionString(connectionString, nil)

	if err != nil {
		log.Debug("Failed to create Service Bus client")
		log.Error("Failed to create Service Bus client", "error", err)
		os.Exit(1)
	} else {
		log.Debug("Service Bus client created")
	}

	// Ensure each topic exists or create it if not
	for _, table := range c.Ingester.Tables {
		topicName, err := servicebus.GenTopicName(c.Ingester.DBConnectionString, table.Name)
		if err != nil {
			log.Error("Failed to generate topic name", "table", table.Name, "error", err)
			os.Exit(1)
		}
		log.Info("Ensuring topic exists", "topic", topicName)

		// Check and create topic if it doesn't exist
		if err := servicebus.CreateTopicIfNotExists(client, topicName); err != nil {
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
	if err != nil {
		// If the queue already exists (409 Conflict), that's fine
		if strings.Contains(err.Error(), "409 Conflict") {
			log.Info("Ingest queue already exists", "queue", c.Ingester.Queue.Name)
		} else {
			log.Error("Failed to create ingest queue", "queue", c.Ingester.Queue.Name, "error", err)
			os.Exit(1)
		}
	} else {
		log.Info("Created ingest queue", "queue", c.Ingester.Queue.Name)
	}
	log.Info("Validated ingest queue", "queue", c.Ingester.Queue.Name)
}

// RenderHCLTemplate applies sprig templating to the HCL file before parsing
func RenderHCLTemplate(filePath string) (string, error) {
	baseName := filepath.Base(filePath)
	tmpl, err := template.New(baseName).Funcs(sprig.TxtFuncMap()).ParseFiles(filePath)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, baseName, nil); err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}

	return buf.String(), nil
}

// RenderHCLTemplateBytes renders a Sprig-enabled HCL template from byte content
func RenderHCLTemplateBytes(name string, content []byte) (string, error) {
	tmpl, err := template.New(name).Funcs(sprig.TxtFuncMap()).Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}

	return buf.String(), nil
}

// RenderHCLTemplate loads the file from disk and renders it via Sprig
func RenderHCLTemplateFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	baseName := filepath.Base(filePath)
	return RenderHCLTemplateBytes(baseName, content)
}

// DecodeHCL returns a config object based on the provided config file
func DecodeHCL[T any](configHCL string, filePath string) T {
	// Parse HCL config starting from position 0
	src := []byte(configHCL)
	pos := hcl.Pos{Line: 0, Column: 0, Byte: 0}
	f, _ := hclsyntax.ParseConfig(src, filePath, pos)

	// Decode HCL into a config struct and return to caller
	var c T
	decodeDiags := gohcl.DecodeBody(f.Body, nil, &c)
	if decodeDiags.HasErrors() {
		log.Error("Error decoding HCL", "error", decodeDiags.Error())
		os.Exit(1)
	}

	return c
}
