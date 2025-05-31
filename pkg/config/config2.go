package config

import "os"

type RootHCL struct {
	DStream *DStreamConfig `hcl:"dstream,block"`
	Tasks   []TaskBlock    `hcl:"task,block"`
}

type DStreamConfig struct {
	Ingest         IngestConfig `hcl:"ingest,block"`
	PluginRegistry string       `hcl:"plugin_registry,attr"`
	Plugins        []PluginSpec `hcl:"required_plugins,block"` // array of objects, not block
}

type IngestConfig struct {
	Provider    string        `hcl:"provider,attr"`
	IngestQueue QueueConfig   `hcl:"ingest_queue,block"`
	Lock        LockConfig    `hcl:"lock,block"`
	Polling     PollingConfig `hcl:"polling,block"`
}

type PollingConfig struct {
	Interval    string `hcl:"interval,attr"`
	MaxInterval string `hcl:"max_interval,attr"`
}

type PluginSpec struct {
	Name    string `hcl:"name"`
	Version string `hcl:"version"`
}

func NewRootHCL(fileName ...string) *RootHCL {

	var configFile string
	if len(fileName) > 0 {
		configFile = fileName[0]
	} else {
		configFile = "dstream.hcl"
	}

	// Load config file
	config, err := LoadRootHCL(configFile)
	if err != nil {
		log.Error("Error loading config", "error", err)
		os.Exit(1)
	}

	return config
}

func LoadRootHCL(filePath string) (*RootHCL, error) {
	var config RootHCL

	// Render HCL config post text templating
	hcl, err := RenderHCLTemplate(filePath)
	if err != nil {
		log.Error("Error generating HCL", "error", err)
		os.Exit(1)
	}

	// Decode HCL to RootHCL struct
	config = DecodeHCL[RootHCL](hcl, filePath)

	return &config, nil
}
