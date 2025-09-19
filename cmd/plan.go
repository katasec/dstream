package cmd

import (
	"fmt"
	"os"

	"github.com/katasec/dstream/pkg/config"
	"github.com/katasec/dstream/pkg/executor"
	"github.com/spf13/cobra"
)

var planCmd = &cobra.Command{
	Use:   "plan [task_name]",
	Short: "Show planned infrastructure changes for a specific task",
	Long: `Show what infrastructure resources would be created or destroyed for a task.

This command will:
- Preview infrastructure changes without making them
- Show what resources would be created by init
- Show what resources would be destroyed by destroy
- Validate provider configurations

Example:
  dstream plan mssql-to-asb    # Show planned changes for mssql-to-asb task

This is similar to 'terraform plan' - it shows what would happen without making changes.`,
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

		// Execute infrastructure planning
		if err := executor.ExecuteTaskWithCommand(task, "plan"); err != nil {
			log.Error("Task planning failed", "task", taskName, "error", err.Error())
			os.Exit(1)
		}

		fmt.Printf("✅ Infrastructure plan for task %q completed successfully\n", taskName)
	},
}

func init() {
	rootCmd.AddCommand(planCmd)
}