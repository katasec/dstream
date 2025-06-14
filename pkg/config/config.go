package config

import (
	"os"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// Config holds the entire configuration as represented in the HCL file
type Config struct {
	Ingesters []LabeledIngester `hcl:"ingester,block"`
	Ingester  Ingester          // temporary active one
	Router    Router            `hcl:"router,block"`
}

type LabeledIngester struct {
	Name                 string              `hcl:",label"`
	DBType               string              `hcl:"db_type,optional"` // for backward compatibility
	DBConnectionString   string              `hcl:"db_connection_string,attr"`
	PollIntervalDefaults PollInterval        `hcl:"poll_interval_defaults,block"`
	Queue                QueueConfig         `hcl:"queue,block"`
	Locks                LockConfig          `hcl:"locks,block"`
	RawTables            []string            `hcl:"tables,attr"`
	TablesOverrides      TableOverridesBlock `hcl:"tables_overrides,block"`
	PluginPath           string              `hcl:"plugin_path,optional"`

	// Computed field
	ResolvedTables []ResolvedTableConfig
}

type Ingester struct {
	DBType               string              `hcl:"db_type,attr"`
	DBConnectionString   string              `hcl:"db_connection_string,attr"`
	PollIntervalDefaults PollInterval        `hcl:"poll_interval_defaults,block"`
	Queue                QueueConfig         `hcl:"queue,block"`
	Locks                LockConfig          `hcl:"locks,block"`
	RawTables            []string            `hcl:"tables,attr"`
	TablesOverrides      TableOverridesBlock `hcl:"tables_overrides,block"`

	Tables []ResolvedTableConfig
}

type ResolvedTableConfig struct {
	Name               string
	PollInterval       string
	MaxPollInterval    string
	DBConnectionString string
	Output             OutputConfig
}

// GetPollInterval returns the PollInterval as a time.Duration
func (t *ResolvedTableConfig) GetPollInterval() (time.Duration, error) {
	return time.ParseDuration(t.PollInterval)
}

// GetMaxPollInterval returns the MaxPollInterval as a time.Duration
func (t *ResolvedTableConfig) GetMaxPollInterval() (time.Duration, error) {
	return time.ParseDuration(t.MaxPollInterval)
}

type QueueConfig struct {
	Type             string `hcl:"type,attr"`              // Type of queue (servicebus, eventhub)
	Name             string `hcl:"name,attr"`              // Name of the queue
	ConnectionString string `hcl:"connection_string,attr"` // Connection string for the queue
}

// LockConfig represents the configuration for distributed locking
type LockConfig struct {
	Type             string `hcl:"type"`                   // Specifies the lock provider type (e.g., "azure_blob")
	ConnectionString string `hcl:"connection_string,attr"` // Connection string for the lock provider
	ContainerName    string `hcl:"container_name"`         // Name of the container used for lock files
}

type PollInterval struct {
	PollInterval    string `hcl:"poll_interval,attr"`
	MaxPollInterval string `hcl:"max_poll_interval,attr"`
}

type TableOverridesBlock struct {
	Overrides []TableOverride `hcl:"overrides,block"`
}

type TableOverride struct {
	TableName       string  `hcl:"table_name,attr"`
	PollInterval    *string `hcl:"poll_interval,optional"`
	MaxPollInterval *string `hcl:"max_poll_interval,optional"`
}

type Router struct {
	Source SourceConfig `hcl:"source,block"` // e.g., "EventHub", "ServiceBus", "Console"
	Output OutputConfig `hcl:"output,block"`
}

// OutputConfig represents the configuration for output type and connection string
type OutputConfig struct {
	Type             string `hcl:"type"`                   // e.g., "EventHub", "ServiceBus", "Console"
	ConnectionString string `hcl:"connection_string,attr"` // Connection string for EventHub or ServiceBus if needed
}

// SourceConfig represents the configuration for the source
type SourceConfig struct {
	Type             string `hcl:"type,attr"`              // e.g., "azure_service_bus"
	ConnectionString string `hcl:"connection_string,attr"` // Connection string for EventHub or ServiceBus if needed
}

func NewConfig(fileName ...string) *Config {

	var configFile string
	if len(fileName) > 0 {
		configFile = fileName[0]
	} else {
		configFile = "dstream.hcl"
	}

	// Load config file
	config, err := LoadConfig(configFile)
	if err != nil {
		log.Error("Error loading config", "error", err)
		os.Exit(1)
	}

	return config
}

// LoadConfig reads, processes the HCL configuration file, and replaces placeholders with environment variables
func LoadConfig(filePath string) (*Config, error) {
	var config Config

	// Generate HCL config post text templating
	hcl, err := RenderHCLTemplate(filePath)
	if err != nil {
		log.Error("Error generating HCL", "error", err)
		os.Exit(1)
	}

	// Read config from generated HCL
	config = processHCL2(hcl, filePath)

	// Get the sqlserver ingester
	for _, ing := range config.Ingesters {
		if ing.Name == "sqlserver" {
			config.Ingester = Ingester{
				DBConnectionString:   ing.DBConnectionString,
				PollIntervalDefaults: ing.PollIntervalDefaults,
				Queue:                ing.Queue,
				Locks:                ing.Locks,
				RawTables:            ing.RawTables,
				TablesOverrides:      ing.TablesOverrides,
				Tables:               nil, // will be filled next
			}
			break
		}
	}

	// Merge defaults and overrides for tables
	resolvedTables := []ResolvedTableConfig{}
	for _, tableName := range config.Ingester.RawTables {
		resolvedTable := ResolvedTableConfig{
			Name:               tableName,
			PollInterval:       config.Ingester.PollIntervalDefaults.PollInterval,
			MaxPollInterval:    config.Ingester.PollIntervalDefaults.MaxPollInterval,
			DBConnectionString: config.Ingester.DBConnectionString,
			Output:             config.Router.Output,
		}

		// Check for overrides
		for _, override := range config.Ingester.TablesOverrides.Overrides {
			if override.TableName == tableName {
				if override.PollInterval != nil {
					resolvedTable.PollInterval = *override.PollInterval
				}
				if override.MaxPollInterval != nil {
					resolvedTable.MaxPollInterval = *override.MaxPollInterval
				}
				break
			}
		}

		resolvedTables = append(resolvedTables, resolvedTable)
	}

	config.Ingester.Tables = resolvedTables

	return &config, nil
}

// processHCL returns a config object based on the provided config file
func processHCL2(configHCL string, filePath string) Config {
	// Parse HCL config starting from position 0
	src := []byte(configHCL)
	pos := hcl.Pos{Line: 0, Column: 0, Byte: 0}
	f, _ := hclsyntax.ParseConfig(src, filePath, pos)

	// Decode HCL into a config struct and return to caller
	var c Config
	decodeDiags := gohcl.DecodeBody(f.Body, nil, &c)
	if decodeDiags.HasErrors() {
		log.Error("Error decoding HCL", "error", decodeDiags.Error())
		os.Exit(1)
	}

	return c
}
