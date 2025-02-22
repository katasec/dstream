package cmd

import (
	"github.com/katasec/dstream/internal/server"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the DStream server",
	Long: `Start the DStream server which monitors SQL Server tables for changes 
using Change Data Capture (CDC) and publishes them to configured destinations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dStream := server.NewServer()
		return dStream.Start()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
