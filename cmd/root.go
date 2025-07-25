package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/katasec/dstream/internal/logging"
	"github.com/spf13/cobra"
)

var (
	cfgFile     string
	logLevel    string
	logFormat   string
	logWithTime bool
)

var rootCmd = &cobra.Command{
	Use:   "dstream",
	Short: "DStream - A plugin-based data streaming tool",
	Long:  `DStream is a plugin-based tool for streaming data between various sources and destinations using a flexible plugin architecture.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logLevel = resolveLogLevel()
		logging.SetLogLevel(logLevel)
		logging.GetHCLogger().Info("Log level set to", "level", logLevel)
	},
}

func init() {
	// Add persistent flags for config and logging
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "dstream.hcl", "Config file path")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "", "Set log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVarP(&logFormat, "log-format", "f", "text", "Set log format (text, json)")
	rootCmd.PersistentFlags().BoolVarP(&logWithTime, "log-time", "t", false, "Include timestamp in logs")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func resolveLogLevel() string {
	if logLevel == "" {
		logLevel = strings.ToLower(os.Getenv("DSTREAM_LOG_LEVEL"))
		if logLevel == "" {
			logLevel = "info" // final fallback
		}
	}

	return logLevel
}
