package config

import (
	"github.com/hashicorp/hcl/v2"
)

// Represents a single `task` block
type TaskBlock struct {
	Type       string   `hcl:"type,label"` // e.g. "ingester"
	Name       string   `hcl:"name,label"` // e.g. "ingest_mssql"
	PluginPath string   `hcl:"plugin_path,optional"`
	PluginRef  string   `hcl:"plugin_ref,optional"`
	Config     hcl.Body `hcl:"config,remain"` // plugin-specific config block
}

// Wrapper for decoding all tasks
type TaskFile struct {
	Tasks []*TaskBlock `hcl:"task,block"`
}
