package cmd

import (
	"fmt"
	"os"

	"github.com/katasec/dstream/pkg/config"
	"github.com/katasec/dstream/pkg/executor"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [task_name]",
	Short: "Show current infrastructure status for a specific task",
	Long: `Show the current state of infrastructure resources for a task.

This command will:
- Display current status of infrastructure resources
- Show which resources exist and their current state
- Report on resource health and availability
- Validate infrastructure configuration

Example:
  dstream status mssql-to-asb    # Show status of mssql-to-asb task infrastructure

This provides visibility into the current state of your streaming infrastructure.`,
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

		// Execute infrastructure status check
		if err := executor.ExecuteTaskWithCommand(task, "status"); err != nil {
			log.Error("Task status check failed", "task", taskName, "error", err.Error())
			os.Exit(1)
		}

		fmt.Printf("✅ Infrastructure status for task %q retrieved successfully\n", taskName)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}