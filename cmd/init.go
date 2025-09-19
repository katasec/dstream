package cmd

import (
	"fmt"
	"os"

	"github.com/katasec/dstream/pkg/config"
	"github.com/katasec/dstream/pkg/executor"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [task_name]",
	Short: "Initialize infrastructure for a specific task",
	Long: `Initialize and provision the required infrastructure for a task.

This command will:
- Create necessary queues, topics, or other infrastructure resources
- Prepare the environment for data streaming
- Validate provider configurations

Example:
  dstream init mssql-to-asb    # Initialize infrastructure for mssql-to-asb task`,
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

		// Execute infrastructure initialization
		if err := executor.ExecuteTaskWithCommand(task, "init"); err != nil {
			log.Error("Task initialization failed", "task", taskName, "error", err.Error())
			os.Exit(1)
		}

		fmt.Printf("✅ Infrastructure for task %q initialized successfully\n", taskName)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}