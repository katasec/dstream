package executor

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-plugin"

	"github.com/katasec/dstream/internal/logging"
	"github.com/katasec/dstream/pkg/config"
	"github.com/katasec/dstream/pkg/orasfetch"
	"github.com/katasec/dstream/pkg/plugins"
	"github.com/katasec/dstream/pkg/plugins/serve"
	pb "github.com/katasec/dstream/proto" // for schema dump
	sdkLogging "github.com/katasec/dstream/sdk/logging"
)

// ExecuteTask looks up a task in dstream.hcl and runs its plugin via gRPC.
func ExecuteTask(task *config.TaskBlock) error {

	// if jsonCfg, err := task.DumpConfigAsJSON(); err != nil {
	// 	log.Error("Failed to dump config as JSON.", "Error:", err.Error())
	// 	return err
	// } else {
	// 	log.Info("Raw interpolated config block:", jsonCfg)
	// }

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

	// Create a custom logger for the plugin client that doesn't add redundant prefixes
	hostLogger := logging.GetHCLogger()

	// Use the clean logger without prefixes
	sdklogger := sdkLogging.SetupCleanLogger() // Initialize clean SDK logging without prefixes
	sdklogger.Info("Hello sdklogger")

	// Set the logger for the plugin client
	//hostLogger.SetLogger(sdklogger)
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: serve.Handshake,
		Plugins: map[string]plugin.Plugin{
			"default": &serve.GenericServerPlugin{},
		},
		Cmd:              cmd,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           sdklogger,
	})

	hostLogger.Info("Hello World!")

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

	// ── kick off the plugin with graceful shutdown support ──────────────────────────────────────────────
	//log.Debug("Starting plugin with config:", cfgStruct)

	// Get input config if available
	inputConfig, err := t.InputAsStructPB()
	if err != nil {
		return fmt.Errorf("decode input config: %w", err)
	}

	// Get output config if available
	outputConfig, err := t.OutputAsStructPB()
	if err != nil {
		return fmt.Errorf("decode output config: %w", err)
	}

	// Create a StartRequest with the config, input, and output
	startReq := &pb.StartRequest{
		Config: cfgStruct,
		Input:  inputConfig,
		Output: outputConfig,
	}

	// Use RunWithGracefulShutdown to handle signals and graceful termination
	err = RunWithGracefulShutdown(context.Background(), func(ctx context.Context) error {
		return pluginClient.Start(ctx, startReq)
	})

	if err != nil {
		return fmt.Errorf("plugin start failed: %w", err)
	}
	return nil
}

// dumpSchema logs the schema in a readable format.
func dumpSchema(fields []*pb.FieldSchema) {
	log.Debug("Schema returned from plugin:")
	for _, f := range fields {
		log.Debug(fmt.Sprintf("- %s (%s) required=%v — %s",
			f.Name, f.Type, f.Required, f.Description))
	}
}
