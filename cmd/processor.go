package cmd

import (
	"github.com/katasec/dstream/internal/config"
	"github.com/katasec/dstream/internal/logging"
	"github.com/katasec/dstream/internal/processor"
	"github.com/spf13/cobra"
)

var log = logging.GetLogger()

var processorCmd = &cobra.Command{
	Use:   "processor",
	Short: "Start the change processor service",
	Long: `Starts the processor service that consumes changes from the queue,
processes them, and publishes them to configured destinations.`,
	Run: runProcessor,
}

func init() {
	rootCmd.AddCommand(processorCmd)
}

func runProcessor(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		log.Error("Failed to load configuration", "error", err)
		return
	}

	// Create and start the processor
	p := processor.NewProcessor(cfg)
	if err := p.Start(); err != nil {
		log.Error("Failed to start processor", "error", err)
		return
	}
}
