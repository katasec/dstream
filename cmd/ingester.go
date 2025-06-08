package cmd

import (
	"fmt"

	"github.com/katasec/dstream/pkg/config"
	"github.com/katasec/dstream/pkg/executor"
	"github.com/katasec/dstream/internal/logging"
	"github.com/spf13/cobra"
)

var (
	taskName string
)

// ingesterCmd spins up any task whose name is supplied with --task (default:
// "ingester-mssql"). It uses the same executor logic as `dstream run`.
var ingesterCmd = &cobra.Command{
	Use:   "ingester",
	Short: "Run an ingester task defined in dstream.hcl",
	Long: `Launches the specified ingester task via the standard plugin
executor. This avoids importing plugin code directly into the CLI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logging.GetHCLogger()

		root, err := config.LoadRootHCL()
		if err != nil {
			return fmt.Errorf("failed to load dstream.hcl: %w", err)
		}

		var task *config.TaskBlock
		for i := range root.Tasks {
			if root.Tasks[i].Name == taskName {
				task = &root.Tasks[i]
				break
			}
		}
		if task == nil {
			return fmt.Errorf("task %q not found in configuration", taskName)
		}

		if err := executor.ExecuteTask(task); err != nil {
			log.Error("Task execution failed", "task", taskName, "error", err)
			return err
		}
		return nil
	},
}

func init() {
	// Default task name mirrors common example but can be overridden.
	ingesterCmd.Flags().StringVar(&taskName, "task", "ingester-mssql",
		"Name of the task block to execute from dstream.hcl")
	rootCmd.AddCommand(ingesterCmd)
}
