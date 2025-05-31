package executor

import (
	"context"
	"fmt"
	"os/exec"

	hplugin "github.com/hashicorp/go-plugin"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/katasec/dstream/pkg/config"
	"github.com/katasec/dstream/pkg/orasfetch"
	"github.com/katasec/dstream/pkg/plugins/serve"
)

// ExecuteTask starts a gRPC plugin using the task config block
func ExecuteTask(task *config.TaskBlock) error {
	fmt.Println("*********** Executing task:", task.Name, "***********")

	// Load the full rendered HCL to extract this task's config block
	rawHCL, err := config.RenderHCLTemplate("dstream.hcl")
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
		HandshakeConfig: serve.Handshake,
		Plugins: map[string]hplugin.Plugin{
			"default": &serve.GenericPlugin{},
		},
		Cmd:              exec.Command(pluginPath),
		AllowedProtocols: []hplugin.Protocol{hplugin.ProtocolGRPC},
	})

	rpcClient, err := client.Client()
	if err != nil {
		return fmt.Errorf("failed to create RPC client: %w", err)
	}

	raw, err := rpcClient.Dispense("default")
	if err != nil {
		log.Error("failed to dispense plugin:", "error", err)
		return fmt.Errorf("failed to dispense plugin: %w", err)
	} else {
		log.Info("Plugin dispensed successfully")
	}

	pluginImpl, ok := raw.(serve.Plugin)
	if !ok {
		return fmt.Errorf("plugin does not implement serve.Plugin interface")
	}

	// Decode config HCL into map[string]string
	hclFile, diags := hclsyntax.ParseConfig(configBytes, "plugin_config.hcl", hcl.InitialPos)
	if diags.HasErrors() {
		return fmt.Errorf("failed to parse config block: %w", diags)
	}

	configMap := make(map[string]string)
	diags = gohcl.DecodeBody(hclFile.Body, nil, &configMap)
	if diags.HasErrors() {
		log.Error("Cannot decode config block, using empty map", "config", configMap)
		return fmt.Errorf("failed to decode config block: %w", diags)
	}

	// Execute the plugin
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := pluginImpl.Start(ctx, configMap); err != nil {
		return fmt.Errorf("plugin execution failed: %w", err)
	}

	return nil
}
