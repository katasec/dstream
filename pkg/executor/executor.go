package executor

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"github.com/katasec/dstream/pkg/config"
	"github.com/katasec/dstream/pkg/logging"
	"github.com/katasec/dstream/pkg/orasfetch"
	"github.com/katasec/dstream/pkg/plugins" // NEW – universal interface
	"github.com/katasec/dstream/pkg/plugins/serve"
	pb "github.com/katasec/dstream/proto" // for schema dump
)

// ExecuteTask looks up a task in dstream.hcl and runs its plugin via gRPC.
func ExecuteTask(task *config.TaskBlock) error {

	if jsonCfg, err := task.DumpConfigAsJSON(); err != nil {
		log.Error("Failed to dump config as JSON.", "Error:", err.Error())
		return err
	} else {
		log.Info("Raw interpolated config block:", jsonCfg)
	}

	// ── reload HCL root (env interpolation may change) ──────────────────
	root, err := config.LoadRootHCL()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// locate the task
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

	// ── decode its `config { … }` block into *structpb.Struct ───────────
	cfgStruct, err := t.ConfigAsStructPB()
	if err != nil {
		return fmt.Errorf("decode config: %w", err)
	}

	// ── resolve plugin binary (local path or pull via ORAS) ─────────────
	var pluginPath string
	switch {
	case t.PluginPath != "":
		pluginPath = t.PluginPath
	case t.PluginRef != "":
		pluginPath, err = orasfetch.PullBinary(t.PluginRef)
		if err != nil {
			return fmt.Errorf("pull plugin: %w", err)
		}
	default:
		return fmt.Errorf("task %q must specify plugin_path or plugin_ref", t.Name)
	}

	// ── launch plugin via HashiCorp go-plugin (gRPC only) ───────────────
	cmd := exec.Command(pluginPath)
	cmd.Stdout, cmd.Stderr = nil, nil

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: serve.Handshake,
		Plugins: map[string]plugin.Plugin{
			"default": &serve.GenericServerPlugin{},
		},
		Cmd:              cmd,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           logging.NewHcLogAdapter(log),
	})

	rpcClient, err := client.Client()
	if err != nil {
		return fmt.Errorf("RPC client setup failed: %w", err)
	}
	raw, err := rpcClient.Dispense("default")

	if err != nil {
		return fmt.Errorf("dispense plugin: %w", err)
	}
	pluginClient := raw.(plugins.Plugin) // ← new universal interface

	// ── optional: dump schema for debugging ──────────────────────────────
	if schema, err := pluginClient.GetSchema(context.Background()); err == nil {
		dumpSchema(schema)
	}

	// ── kick off the plugin ──────────────────────────────────────────────
	log.Debug("Starting plugin with config:", cfgStruct)
	if err := pluginClient.Start(context.Background(), cfgStruct); err != nil {
		return fmt.Errorf("plugin start failed: %w", err)
	}
	return nil
}

// dumpSchema logs the schema in a readable format.
func dumpSchema(fields []*pb.FieldSchema) {
	log.Info("Schema returned from plugin:")
	for _, f := range fields {
		log.Info(fmt.Sprintf("- %s (%s) required=%v — %s",
			f.Name, f.Type, f.Required, f.Description))
	}
}
