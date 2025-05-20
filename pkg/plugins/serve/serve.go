// pkg/plugins/serve.go
package serve

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/katasec/dstream/pkg/plugins"
	"github.com/katasec/dstream/pkg/plugins/ingester"
	pb "github.com/katasec/dstream/proto"
	"google.golang.org/grpc"
)

// HandshakeConfig for verifying CLI ↔ plugin compatibility
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "DSTREAM_PLUGIN",
	MagicCookieValue: "hello",
}

// IngesterPlugin is a HashiCorp go-plugin wrapper for gRPC-based ingesters
type IngesterPlugin struct {
	plugin.Plugin
	Impl plugins.Ingester
}

func (p *IngesterPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	pb.RegisterIngesterPluginServer(s, &ingester.GRPCServer{Impl: p.Impl})
	return nil
}

func (p *IngesterPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	return ingester.NewGRPCClient(pb.NewIngesterPluginClient(conn)), nil
}
