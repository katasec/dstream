package logging

import (
	"os"

	"github.com/hashicorp/go-hclog"
)

var (
	// Environment variable names
	envLogLevel = "DSTREAM_LOG_LEVEL" // debug, info, warn, error
	envLogJSON  = "DSTREAM_LOG_JSON"  // true/false
)

// GetLogger returns an hclog-compatible logger for plugins to use
// This logger does NOT add timestamps as plugins should let the host
// handle timestamp formatting to avoid duplication
func GetLogger(name string) hclog.Logger {
	// Get log level from environment
	logLevel := hclog.Info
	if level := os.Getenv(envLogLevel); level != "" {
		logLevel = hclog.LevelFromString(level)
	}

	// Check if JSON logging is enabled
	jsonFormat := false
	if jsonEnv := os.Getenv(envLogJSON); jsonEnv == "true" || jsonEnv == "1" {
		jsonFormat = true
	}

	// Create a logger with no timestamps - the host will add these
	return hclog.New(&hclog.LoggerOptions{
		Name:             name,
		Level:            logLevel,
		Output:           os.Stderr, // Important: use stderr for plugin logs
		JSONFormat:       jsonFormat,
		DisableTime:      true, // Don't add timestamps - host will handle this
		IncludeLocation:  false,
		IndependentLevels: true,
		ColorHeaderOnly:  true,
	})
}

// SetupPluginLogger is a convenience function for plugin main functions
// It creates a logger and returns it, using the plugin name as the logger name
func SetupPluginLogger(pluginName string) hclog.Logger {
	return GetLogger(pluginName)
}
