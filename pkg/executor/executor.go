package executor

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"github.com/katasec/dstream/pkg/config"
	"github.com/katasec/dstream/pkg/orasfetch"
	"github.com/katasec/dstream/pkg/plugins/serve"
)

// ExecuteTask starts a gRPC plugin using the task config block
func ExecuteTask(task *config.TaskBlock) error {
	log.Info("*********** Executing task:", task.Name, "***********")

	rawHCL, err := config.RenderHCLTemplate("dstream.hcl")
	if err != nil {
		return fmt.Errorf("failed to read HCL file: %w", err)
	}

	configBytes, err := config.ExtractConfigBlock(string(rawHCL), task.Type, task.Name)
	if err != nil {
		log.Error("Cannot decode config block, using empty map", "config", configBytes)
		return fmt.Errorf("failed to decode config block: %w", err)
	}

	log.Info("Executor:", "Name:", task.Name, "Type:", task.Type)

	// Pull plugin
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

	// Start plugin
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: serve.Handshake,
		Plugins: map[string]plugin.Plugin{
			"default": &serve.GenericPlugin{},
		},
		Cmd:              exec.Command(pluginPath),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})

	rpcClient, err := client.Client()
	if err != nil {
		return fmt.Errorf("failed to create RPC client: %w", err)
	}

	rawPlugin, err := rpcClient.Dispense("default")
	if err != nil {
		return fmt.Errorf("failed to dispense plugin: %w", err)
	}

	// ✅ Use the serve.Plugin interface, not proto.PluginClient
	pluginClient := rawPlugin.(serve.Plugin)

	// ✅ GetSchema using serve.Plugin signature
	schemaFields, err := pluginClient.GetSchema(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get plugin schema: %w", err)
	}

	log.Info("Schema returned from plugin:")
	for _, field := range schemaFields {
		log.Info(fmt.Sprintf("- %s (%s): %s [required=%v]",
			field.Name, field.Type, field.Description, field.Required))
	}

	// Skip Start for now, or implement below
	// err = pluginClient.Start(context.Background(), configBytes)
	// if err != nil {
	//     return fmt.Errorf("plugin start failed: %w", err)
	// }

	return nil
}
