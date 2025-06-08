// Package logging provides SDK logging functions for plugins.
// This file contains functions specifically for plugin developers to use.
package logging

import (
	"github.com/hashicorp/go-hclog"
	sdkLogging "github.com/katasec/dstream/sdk/logging"
)

// GetPluginLogger returns an hclog-compatible logger for plugins.
// This logger is specifically designed for plugin use and does NOT add timestamps,
// as the host process will handle adding timestamps when it processes the logs.
// 
// Plugin developers should use this function to get a logger for their plugins.
func GetPluginLogger(name string) hclog.Logger {
	return sdkLogging.GetLogger(name)
}

// SetupPluginLogger is a convenience function for plugin main functions.
// It creates a logger with the plugin name and returns it.
// 
// This is the recommended way for plugins to set up their logging in main().
func SetupPluginLogger(pluginName string) hclog.Logger {
	return sdkLogging.SetupPluginLogger(pluginName)
}
