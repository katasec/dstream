package config

import "os"

type RootHCL struct {
	Locals  *LocalsBlock   `hcl:"locals,block"`
	DStream *DStreamConfig `hcl:"dstream,block"`
	Tasks   []TaskBlock    `hcl:"task,block"`
}

type LocalsBlock struct {
	Vars map[string]interface{} `hcl:",remain"`
}

type DStreamConfig struct {
	PluginRegistry string       `hcl:"plugin_registry,attr"`
	Plugins        []PluginSpec `hcl:"required_plugins,block"` // array of objects, not block
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

func LoadRootHCL(fileName ...string) (*RootHCL, error) {

	// Get optional file name, default to "dstream.hcl"
	var configFile string
	if len(fileName) > 0 {
		configFile = fileName[0]
	} else {
		configFile = "dstream.hcl"
	}

	var config RootHCL

	// Render HCL config post text templating
	hcl, err := RenderHCLTemplate(configFile)
	if err != nil {
		log.Error("Error generating HCL", "error", err)
		os.Exit(1)
	}

	// Decode HCL to RootHCL struct
	config = DecodeHCL[RootHCL](hcl, configFile)

	return &config, nil
}
