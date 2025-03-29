package cmd

import (
	"fmt"

	"github.com/katasec/dstream/internal/ingester"
	"github.com/spf13/cobra"
)

var ingesterCmd = &cobra.Command{
	Use:   "ingester",
	Short: "Start the DStream ingester",
	Long: `Start the DStream ingester which monitors SQL Server tables for changes 
using Change Data Capture (CDC) and sends them to the ingest queue.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dStream, err := ingester.NewIngester()
		if err != nil {
			return fmt.Errorf("failed to create ingester: %w", err)
		}
		return dStream.Start()
	},
}

func init() {
	rootCmd.AddCommand(ingesterCmd)
}
