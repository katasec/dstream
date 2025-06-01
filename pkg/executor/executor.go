package executor

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"github.com/katasec/dstream/pkg/config"
	"github.com/katasec/dstream/pkg/logging"
	"github.com/katasec/dstream/pkg/orasfetch"
	"github.com/katasec/dstream/pkg/plugins/serve"
)

// ExecuteTask looks up the task in dstream.hcl and runs its plugin via gRPC.
func ExecuteTask(task *config.TaskBlock) error {
	log.Info("*********** Executing task:", task.Name, "***********")

	//----------------------------------------------------------------------
	// Load the root HCL once
	//----------------------------------------------------------------------
	root, err := config.LoadRootHCL()
	if err != nil {
		log.Error("Failed to load configuration", "error", err)
		return err
	}
	log.Debug("Configuration loaded successfully")

	//----------------------------------------------------------------------
	// Locate the requested task block
	//----------------------------------------------------------------------
	var t *config.TaskBlock
	for i := range root.Tasks {
		if root.Tasks[i].Name == task.Name {
			t = &root.Tasks[i]
			break
		}
	}
	if t == nil {
		return fmt.Errorf("task %q not found in configuration", task.Name)
	}
	log.Debug("Found task in configuration", "task", t.Name)

	//----------------------------------------------------------------------
	// Decode its <config> block into map[string]string
	//----------------------------------------------------------------------
	typedConfig, _, err := t.ConfigAsStringMap()
	if err != nil {
		return fmt.Errorf("failed to decode config block: %w", err)
	}

	//----------------------------------------------------------------------
	// Resolve the plugin binary (local path or pull via ORAS)
	//----------------------------------------------------------------------
	var pluginPath string
	switch {
	case t.PluginPath != "":
		pluginPath = t.PluginPath

	case t.PluginRef != "":
		pluginPath, err = orasfetch.PullBinary(t.PluginRef)
		if err != nil {
			return fmt.Errorf("failed to pull plugin: %w", err)
		}

	default:
		return fmt.Errorf("task %q must specify either plugin_path or plugin_ref", t.Name)
	}

	log.Info("Executor:", "Launching plugin:", pluginPath)

	// Setup cmd
	cmd := exec.Command(pluginPath)
	cmd.Stdout = nil // silence stdout
	cmd.Stderr = nil // silence stderr

	//----------------------------------------------------------------------
	// Start the plugin via go-plugin
	//----------------------------------------------------------------------
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: serve.Handshake,
		Plugins: map[string]plugin.Plugin{
			"default": &serve.GenericPlugin{},
		},
		Cmd:              cmd,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           logging.NewHcLogAdapter(logging.GetLogger()),
	})

	rpcClient, err := client.Client()
	if err != nil {
		return fmt.Errorf("failed to create RPC client: %w", err)
	}
	rawPlugin, err := rpcClient.Dispense("default")
	if err != nil {
		return fmt.Errorf("failed to dispense plugin: %w", err)
	}
	pluginClient := rawPlugin.(serve.Plugin)

	//----------------------------------------------------------------------
	// Optional: print schema for debugging
	//----------------------------------------------------------------------
	if schema, err := pluginClient.GetSchema(context.Background()); err == nil {
		log.Info("Schema returned from plugin:")
		for _, f := range schema {
			log.Info(fmt.Sprintf("- %s (%s): %s [required=%v]",
				f.Name, f.Type, f.Description, f.Required))
		}
	} else {
		log.Warn("Could not retrieve schema:", err)
	}

	//----------------------------------------------------------------------
	// Kick off the plugin
	//----------------------------------------------------------------------
	if err := pluginClient.Start(context.Background(), typedConfig); err != nil {
		return fmt.Errorf("plugin start failed: %w", err)
	}

	return nil
}
