package serve

import (
	"context"

	hplugin "github.com/hashicorp/go-plugin"
	pb "github.com/katasec/dstream/proto"
	"google.golang.org/grpc"
)

// Plugin is the standard interface all DStream plugins must implement.
type Plugin interface {
	Start(ctx context.Context, config map[string]string) error
}

// GenericPlugin implements the go-plugin plumbing to wire gRPC server + client
type GenericPlugin struct {
	hplugin.Plugin
	Impl Plugin
}

func (p *GenericPlugin) GRPCServer(broker *hplugin.GRPCBroker, s *grpc.Server) error {
	pb.RegisterDStreamPluginServer(s, NewGRPCServer(p.Impl))
	return nil
}

func (p *GenericPlugin) GRPCClient(ctx context.Context, broker *hplugin.GRPCBroker, cc *grpc.ClientConn) (interface{}, error) {
	return NewGRPCClient(cc), nil
}
