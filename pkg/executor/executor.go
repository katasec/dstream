package executor

import (
	"fmt"
	"os/exec"

	"github.com/katasec/dstream/pkg/config"
)

// ExecuteTask starts the plugin defined by the given TaskBlock.
func ExecuteTask(task *config.TaskBlock) error {
	// Step 1: Load raw HCL file
	rawHCL, err := config.GenerateHCL("dstream2.hcl")
	if err != nil {
		return fmt.Errorf("failed to read HCL file: %w", err)
	}

	// Step 2: Extract the config { } block from the task
	configBytes, err := config.ExtractConfigBlock(string(rawHCL), task.Type, task.Name)
	if err != nil {
		return fmt.Errorf("failed to extract config block: %w", err)
	}

	// TEMP: Just print the config to verify
	fmt.Printf("[executor] Launching plugin: %s\n", task.PluginPath)
	fmt.Printf("[executor] Task: %s | Type: %s\n", task.Name, task.Type)
	fmt.Printf("[executor] Raw config:\n%s\n", string(configBytes))

	// Step 3: Start plugin process
	cmd := exec.Command(task.PluginPath)

	// NOTE: Commented out until you wire up gRPC
	/*
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return fmt.Errorf("failed to get stdin pipe: %w", err)
		}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("failed to get stdout pipe: %w", err)
		}
	*/

	// TEMP: Run the plugin without I/O plumbing
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start plugin: %w", err)
	} else {
		fmt.Printf("[executor] Plugin started with PID: %d\n", cmd.Process.Pid)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("plugin exited with error: %w", err)
	} else {
		fmt.Printf("[executor] Plugin exited successfully\n")
	}

	return nil
}
