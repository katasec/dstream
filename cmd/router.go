package cmd

import (
	"github.com/katasec/dstream/internal/router"
	"github.com/katasec/dstream/pkg/config"
	"github.com/katasec/dstream/pkg/logging"
	"github.com/spf13/cobra"
)

var log = logging.GetLogger()

var routerCmd = &cobra.Command{
	Use:   "router",
	Short: "Start the change router service",
	Long: `Starts the router service that consumes changes from the ingest queue,
and routes them to their configured destination topics.`,
	Run: runRouter,
}

func init() {
	rootCmd.AddCommand(routerCmd)
}

func runRouter(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		log.Error("Failed to load configuration", "error", err)
		return
	}

	// Create and start the router
	r, err := router.NewRouter(cfg)
	if err != nil {
		log.Error("Failed to create router", "error", err)
		return
	}

	if err := r.Start(); err != nil {
		log.Error("Failed to start router", "error", err)
		return
	}
}
