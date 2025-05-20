package config

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
