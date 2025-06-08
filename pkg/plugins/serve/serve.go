package serve

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/katasec/dstream/pkg/plugins"
)

// Handshake shared by CLI and all plugins.
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "DSTREAM_PLUGIN",
	MagicCookieValue: "dstream",
}

// Serve is called by a plugin's main() to expose its implementation.
func Serve(impl plugins.Plugin, logger hclog.Logger) {
	// Create a plugin serve config
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins: map[string]plugin.Plugin{
			"default": &GenericServerPlugin{Impl: impl},
		},
		GRPCServer: plugin.DefaultGRPCServer,
		Logger:     logger,
	})
}
