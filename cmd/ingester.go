package cmd

import (
	"context"
	"fmt"

	mssql "github.com/katasec/dstream-ingester-mssql/mssql"
	"github.com/katasec/dstream/pkg/logging"
	"github.com/katasec/dstream/pkg/plugins"
	"github.com/spf13/cobra"
)

var ingesterCmd = &cobra.Command{
	Use:   "ingester",
	Short: "Start the DStream ingester",
	Long:  `Start the DStream ingester which monitors SQL Server tables for changes using Change Data Capture (CDC) and sends them to the ingest queue.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := logging.GetLogger()
		ing := mssql.New()
		ctx := context.Background()

		err := ing.Start(ctx, func(e plugins.Event) error {
			logger.Info("Received event: %+v", e)
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to start ingester: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(ingesterCmd)
}
