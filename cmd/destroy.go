package cmd

import (
	"fmt"
	"os"

	"github.com/katasec/dstream/pkg/config"
	"github.com/katasec/dstream/pkg/executor"
	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy [task_name]",
	Short: "Destroy infrastructure for a specific task",
	Long: `Destroy and clean up the infrastructure resources for a task.

This command will:
- Remove queues, topics, or other infrastructure resources
- Clean up any resources created by the init command
- Ensure no lingering resources remain

Example:
  dstream destroy mssql-to-asb    # Destroy infrastructure for mssql-to-asb task

Warning: This operation is irreversible and will delete infrastructure resources.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskName := args[0]
		hclPath := "dstream.hcl"

		if _, err := os.Stat(hclPath); os.IsNotExist(err) {
			log.Error("Config file not found: %s", hclPath)
			os.Exit(1)
		}

		root, err := config.LoadRootFile(hclPath)
		if err != nil {
			log.Error("Failed to load root config from %s: %v", hclPath, err)
			os.Exit(1)
		}

		var task *config.TaskBlock
		for _, t := range root.Tasks {
			if t.Name == taskName {
				task = &t
				break
			}
		}

		if task == nil {
			log.Error("Task %q not found in %s", taskName, hclPath)
			os.Exit(1)
		}

		// Execute infrastructure destruction
		if err := executor.ExecuteTaskWithCommand(task, "destroy"); err != nil {
			log.Error("Task destruction failed", "task", taskName, "error", err.Error())
			os.Exit(1)
		}

		fmt.Printf("✅ Infrastructure for task %q destroyed successfully\n", taskName)
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}