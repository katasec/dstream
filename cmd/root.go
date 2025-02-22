package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	logLevel    string
	logFormat   string
	logWithTime bool
)

var rootCmd = &cobra.Command{
	Use:   "dstream",
	Short: "DStream - A data streaming tool",
	Long: `DStream is a tool for streaming data changes from SQL Server 
using Change Data Capture (CDC) and publishing them to various destinations.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set environment variables for logging
		os.Setenv("DSTREAM_LOG_LEVEL", logLevel)
		os.Setenv("DSTREAM_LOG_HANDLER", logFormat)
		if logWithTime {
			os.Setenv("DSTREAM_LOG_WITH_TIME", "true")
		}
	},
}

func init() {
	// Add persistent flags for logging
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Set log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVarP(&logFormat, "log-format", "f", "text", "Set log format (text, json)")
	rootCmd.PersistentFlags().BoolVarP(&logWithTime, "log-time", "t", false, "Include timestamp in logs")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
