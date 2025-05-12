package cmd

import (
	"fmt"
	"os"

	"github.com/katasec/dstream/pkg/config"
	"github.com/katasec/dstream/pkg/executor"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [task_name]",
	Short: "Execute a specific task by name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskName := args[0]
		hclPath := "dstream2.hcl"

		if _, err := os.Stat(hclPath); os.IsNotExist(err) {
			log.Error("Config file not found: %s", hclPath)
			os.Exit(1)
		}

		taskFile, err := config.LoadTaskFile(hclPath)
		if err != nil {
			log.Error("Failed to load tasks from %s: %v", hclPath, err)
			os.Exit(1)
		}

		var task *config.TaskBlock
		for _, t := range taskFile.Tasks {
			if t.Name == taskName {
				task = t
				break
			}
		}
		if task == nil {
			log.Error("Task %q not found in %s", taskName, hclPath)
			os.Exit(1)
		}

		if err := executor.ExecuteTask(task); err != nil {
			log.Error("Task execution failed", "task", taskName, "error", err.Error())
			os.Exit(1)
		}

		fmt.Printf("✅ Task %q executed successfully\n", taskName)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
