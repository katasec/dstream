package executor

import (
	"context"
	"fmt"
	"os/exec"

	hplugin "github.com/hashicorp/go-plugin"
	"github.com/katasec/dstream/pkg/config"
	"github.com/katasec/dstream/pkg/orasfetch"
	"github.com/katasec/dstream/pkg/plugins"
	"github.com/katasec/dstream/pkg/plugins/serve"
)

// ExecuteTask starts the plugin defined by the given TaskBlock.
func ExecuteTask(task *config.TaskBlock) error {
	// Load the full HCL file so we can extract the raw config block for this task
	rawHCL, err := config.GenerateHCL("dstream.hcl")
	if err != nil {
		return fmt.Errorf("failed to read HCL file: %w", err)
	}

	configBytes, err := config.ExtractConfigBlock(string(rawHCL), task.Type, task.Name)
	if err != nil {
		return fmt.Errorf("failed to extract config block: %w", err)
	}
	if configBytes != nil {
		log.Info("Extracted config block")
	}

	log.Info("Executor:", "Name:", task.Name, "Type:", task.Type)

	// Resolve plugin path
	var pluginPath string
	if task.PluginPath != "" {
		pluginPath = task.PluginPath
	} else if task.PluginRef != "" {
		pluginPath, err = orasfetch.PullBinary(task.PluginRef)
		if err != nil {
			return fmt.Errorf("failed to pull plugin: %w", err)
		}
	} else {
		return fmt.Errorf("task must have either plugin_path or plugin_ref")
	}

	log.Info("Executor:", "Launching plugin:", pluginPath)

	// Start the plugin client
	client := hplugin.NewClient(&hplugin.ClientConfig{
		HandshakeConfig:  serve.Handshake,
		Plugins:          map[string]hplugin.Plugin{"ingester": &serve.IngesterPlugin{}},
		Cmd:              exec.Command(pluginPath),
		AllowedProtocols: []hplugin.Protocol{hplugin.ProtocolGRPC},
	})

	rpcClient, err := client.Client()
	if err != nil {
		return fmt.Errorf("failed to create RPC client: %w", err)
	}

	raw, err := rpcClient.Dispense("ingester")
	if err != nil {
		log.Error("failed to dispense plugin:", "error", err)
		return fmt.Errorf("failed to dispense plugin: %w", err)
	}

	ingester := raw.(plugins.Ingester)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	emit := func(e plugins.Event) error {
		log.Info("Plugin Emit", "Data", e)
		return nil
	}

	if err := ingester.Start(ctx, emit); err != nil {
		return fmt.Errorf("plugin execution failed: %w", err)
	}

	return nil
}
