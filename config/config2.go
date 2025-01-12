package config

import (
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// Config holds the entire configuration as represented in the HCL file
type Config2 struct {
	Ingester  Ingester  `hcl:"ingester,block"`
	Publisher Publisher `hcl:"publisher,block"`
}

type Ingester struct {
	DBType               string              `hcl:"db_type,attr"`
	DBConnectionString   string              `hcl:"db_connection_string,attr"`
	PollIntervalDefaults PollInterval        `hcl:"poll_interval_defaults,block"`
	Topic                TopicConfig         `hcl:"topic,block"`
	Locks                LockConfig          `hcl:"locks,block"`
	RawTables            []string            `hcl:"tables,attr"`
	TablesOverrides      TableOverridesBlock `hcl:"tables_overrides,block"`

	Tables []ResolvedTableConfig
}

type ResolvedTableConfig struct {
	Name            string
	PollInterval    string
	MaxPollInterval string
}

type TopicConfig struct {
	Name             string `hcl:"name,attr"`
	ConnectionString string `hcl:"connection_string,attr"`
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

type Publisher struct {
	Source SourceConfig `hcl:"source,block"` // e.g., "EventHub", "ServiceBus", "Console"
	Output OutputConfig `hcl:"output,block"`
}

// OutputConfig represents the configuration for output type and connection string
type SourceConfig struct {
	Type             string `hcl:"type,attr"`              // e.g., "azure_service_bus"
	ConnectionString string `hcl:"connection_string,attr"` // Connection string for EventHub or ServiceBus if needed
}

func NewConfig2() *Config2 {
	// Load config file
	config, err := LoadConfig2("dstream.hcl")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	return config
}

// LoadConfig reads, processes the HCL configuration file, and replaces placeholders with environment variables
func LoadConfig2(filePath string) (*Config2, error) {
	var config Config2

	// Generate HCL config post text templating
	hcl, err := generateHCL(filePath)
	if err != nil {
		log.Fatal(err)
	}

	// Read config from generated HCL
	config = processHCL2(hcl, filePath)

	// Merge defaults and overrides for tables
	resolvedTables := []ResolvedTableConfig{}
	for _, tableName := range config.Ingester.RawTables {
		resolvedTable := ResolvedTableConfig{
			Name:            tableName,
			PollInterval:    config.Ingester.PollIntervalDefaults.PollInterval,
			MaxPollInterval: config.Ingester.PollIntervalDefaults.MaxPollInterval,
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
func processHCL2(configHCL string, filePath string) Config2 {
	// Parse HCL config starting from position 0
	src := []byte(configHCL)
	pos := hcl.Pos{Line: 0, Column: 0, Byte: 0}
	f, _ := hclsyntax.ParseConfig(src, filePath, pos)

	// Decode HCL into a config struct and return to caller
	var c Config2
	decodeDiags := gohcl.DecodeBody(f.Body, nil, &c)
	if decodeDiags.HasErrors() {
		log.Fatal(decodeDiags.Error())
	}

	return c
}
