package config

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
	"github.com/Masterminds/sprig"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	queues "github.com/katasec/dstream/queues"
)

// TableConfig represents individual table configurations in the HCL file
type TableConfig struct {
	Name            string `hcl:"name"`
	PollInterval    string `hcl:"poll_interval"`
	MaxPollInterval string `hcl:"max_poll_interval"`
}

// OutputConfig represents the configuration for output type and connection string
type OutputConfig struct {
	Type             string `hcl:"type"`                   // e.g., "EventHub", "ServiceBus", "Console"
	ConnectionString string `hcl:"connection_string,attr"` // Connection string for EventHub or ServiceBus if needed
}

// Config holds the entire configuration as represented in the HCL file
type Config struct {
	DBType             string        `hcl:"db_type"`
	DBConnectionString string        `hcl:"db_connection_string"`
	Output             OutputConfig  `hcl:"output,block"`
	Tables             []TableConfig `hcl:"tables,block"`
}

// CheckConfig validates the configuration based on the output type requirements
func (c *Config) CheckConfig() {
	if c.DBConnectionString == "" {
		log.Println("Error, DBConnectionString was not found, exiting.")
		os.Exit(0)
	}

	// Validate based on Output Type requirements
	switch strings.ToLower(c.Output.Type) {
	case "eventHub":
		if c.Output.ConnectionString == "" {
			log.Fatalf("Error, %s connection string is required.", c.Output.Type)
		}
	case "servicebus":
		c.serviceBusConfigCheck()
	case "console":
		// Console output type doesn't need a connection string.
		log.Println("Output set to console; no additional connection string required.")
	default:
		log.Fatalf("Error, unknown output type: %s", c.Output.Type)
	}
}

// serviceBusConfigCheck validates the Service Bus configuration and ensures queues exist for each table
func (c *Config) serviceBusConfigCheck() {
	if c.Output.ConnectionString == "" {
		log.Fatalf("Error, %s connection string is required.", c.Output.Type)
	}

	// Create a Service Bus admin client
	client, err := admin.NewClientFromConnectionString(c.Output.ConnectionString, nil)
	if err != nil {
		log.Fatalf("Failed to create Service Bus client: %v", err)
	}

	// Ensure each queue exists or create it if not
	for _, table := range c.Tables {
		//queueName := GetServiceBusQueueName(table.Name)
		queueName := queues.GetServiceBusQueueName(c.DBConnectionString, table.Name)
		log.Printf("Ensuring queue exists: %s\n", queueName)

		// Check and create queue if it doesn't exist
		if err := createQueueIfNotExists(client, queueName); err != nil {
			log.Fatalf("Error ensuring queue %s exists: %v", queueName, err)
		}
	}
}

// createQueueIfNotExists checks if a queue exists and creates it if it doesn’t
func createQueueIfNotExists(client *admin.Client, queueName string) error {
	// Check if the queue exists
	response, err := client.GetQueue(context.TODO(), queueName, nil)
	if err == nil && response != nil {
		log.Printf("Queue %s already exists.\n", queueName)
		return nil // Queue already exists
	}
	// Check if the error is a "not found" error based on HTTP status code
	log.Printf("Queue %s does not exist. Creating...\n", queueName)
	response2, err := client.CreateQueue(context.TODO(), queueName, nil)
	if err != nil {
		log.Fatalf("failed to create queue %s: %s", queueName, err)
	}
	fmt.Printf("Queue %s created successfully. Status: %d\n", queueName, response2.Status)
	return nil
}

func GetServiceBusQueueName(tableName string) string {
	tableName = strings.ToLower(tableName)
	return fmt.Sprintf("%s-events", tableName)
}

// LoadConfig reads, processes the HCL configuration file, and replaces placeholders with environment variables
func LoadConfig(filePath string) (*Config, error) {
	var config Config

	// Generate HCL config post text templating
	hcl, err := generateHCL(filePath)
	if err != nil {
		log.Fatal(err)
	}

	// Read config from generated HCL
	config = processHCL(hcl, filePath)

	return &config, nil
}

// GetPollInterval returns the PollInterval as a time.Duration
func (t *TableConfig) GetPollInterval() (time.Duration, error) {
	return time.ParseDuration(t.PollInterval)
}

// GetMaxPollInterval returns the MaxPollInterval as a time.Duration
func (t *TableConfig) GetMaxPollInterval() (time.Duration, error) {
	return time.ParseDuration(t.MaxPollInterval)
}

// generateHCL Generates the HCL config after processing the text templating
func generateHCL(filePath string) (hcl string, err error) {
	// Get the Sprig function map
	fmap := sprig.TxtFuncMap()

	// Define template for *.hcl and *.tpl files in the current folder
	// Ensure the Sprig functions are loaded for processing templates
	t := template.Must(template.New("test").
		Funcs(fmap).
		ParseFiles(filePath))

	buf := &bytes.Buffer{}
	err = t.ExecuteTemplate(buf, filePath, nil)
	if err != nil {
		fmt.Printf("Error during template execution: %s", err)
		return "", err
	}

	return buf.String(), nil
}

// processHCL returns a config object based on the provided config file
func processHCL(configHCL string, filePath string) (config Config) {
	// Parse HCL config starting from position 0
	src := []byte(configHCL)
	pos := hcl.Pos{Line: 0, Column: 0, Byte: 0}
	f, _ := hclsyntax.ParseConfig(src, filePath, pos)

	// Decode HCL into a config struct and return to caller
	var c Config
	decodeDiags := gohcl.DecodeBody(f.Body, nil, &c)
	if decodeDiags.HasErrors() {
		log.Fatal(decodeDiags.Error())
	}

	return c
}
