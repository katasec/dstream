package executor

import (
	"context"
	"fmt"
	"os/exec"

	hplugin "github.com/hashicorp/go-plugin"
	"github.com/katasec/dstream/pkg/config"
	"github.com/katasec/dstream/pkg/plugins"
	"github.com/katasec/dstream/pkg/plugins/serve"
)

// // ExecuteTask starts the plugin defined by the given TaskBlock.
// func ExecuteTask(task *config.TaskBlock) error {
// 	// Step 1: Load raw HCL file
// 	rawHCL, err := config.GenerateHCL("dstream.hcl")
// 	if err != nil {
// 		return fmt.Errorf("failed to read HCL file: %w", err)
// 	}

// 	// Step 2: Extract the config { } block from the task
// 	configBytes, err := config.ExtractConfigBlock(string(rawHCL), task.Type, task.Name)
// 	if err != nil {
// 		return fmt.Errorf("failed to extract config block: %w", err)
// 	}

// 	// TEMP: Just print the config to verify
// 	fmt.Printf("[executor] Launching plugin: %s\n", task.PluginPath)
// 	fmt.Printf("[executor] Task: %s | Type: %s\n", task.Name, task.Type)
// 	fmt.Printf("[executor] Raw config:\n%s\n", string(configBytes))

// 	// Step 3: Start plugin process
// 	cmd := exec.Command(task.PluginPath)

// 	// NOTE: Commented out until you wire up gRPC
// 	/*
// 		stdin, err := cmd.StdinPipe()
// 		if err != nil {
// 			return fmt.Errorf("failed to get stdin pipe: %w", err)
// 		}
// 		stdout, err := cmd.StdoutPipe()
// 		if err != nil {
// 			return fmt.Errorf("failed to get stdout pipe: %w", err)
// 		}
// 	*/

// 	// TEMP: Run the plugin without I/O plumbing
// 	if err := cmd.Start(); err != nil {
// 		fmt.Printf("[executor] Failed to start plugin %s: %v\n", task.PluginPath, err)
// 		return fmt.Errorf("failed to start plugin: %w", err)
// 	} else {
// 		fmt.Printf("[executor] Plugin started with PID: %d\n", cmd.Process.Pid)
// 	}
// 	if err := cmd.Wait(); err != nil {
// 		return fmt.Errorf("plugin exited with error: %w", err)
// 	} else {
// 		fmt.Printf("[executor] Plugin exited successfully\n")
// 	}

// 	return nil
// }

func ExecuteTask(task *config.TaskBlock) error {
	// Generate raw config block
	rawHCL, err := config.GenerateHCL("dstream.hcl")
	if err != nil {
		return fmt.Errorf("failed to read HCL file: %w", err)
	}

	configBytes, err := config.ExtractConfigBlock(string(rawHCL), task.Type, task.Name)
	if err != nil {
		return fmt.Errorf("failed to extract config block: %w", err)
	}

	fmt.Printf("[executor] Launching plugin: %s\n", task.PluginPath)
	fmt.Printf("[executor] Task: %s | Type: %s\n", task.Name, task.Type)
	fmt.Printf("[executor] Raw config:\n%s\n", string(configBytes))

	// Launch gRPC plugin
	client := hplugin.NewClient(&hplugin.ClientConfig{
		HandshakeConfig:  serve.Handshake,
		Plugins:          map[string]hplugin.Plugin{"ingester": &serve.IngesterPlugin{}},
		Cmd:              exec.Command(task.PluginPath),
		AllowedProtocols: []hplugin.Protocol{hplugin.ProtocolGRPC},
	})

	rpcClient, err := client.Client()
	if err != nil {
		return fmt.Errorf("failed to create RPC client: %w", err)
	}

	raw, err := rpcClient.Dispense("ingester")
	if err != nil {
		return fmt.Errorf("failed to dispense plugin: %w", err)
	}

	ingester := raw.(plugins.Ingester)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	emit := func(e plugins.Event) error {
		fmt.Printf("[plugin emit] %+v\n", e)
		return nil
	}

	if err := ingester.Start(ctx, emit); err != nil {
		return fmt.Errorf("plugin execution failed: %w", err)
	}

	return nil
}
