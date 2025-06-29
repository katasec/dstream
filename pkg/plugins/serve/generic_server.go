package serve

import (
	"context"

	"github.com/hashicorp/go-plugin"
	pb "github.com/katasec/dstream/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/katasec/dstream/pkg/plugins"
)

// ───────────────────────────────────────────────────────────────────────────────
//  gRPC service implementation (runs inside the plugin process)
// ───────────────────────────────────────────────────────────────────────────────

type server struct {
	pb.UnimplementedPluginServer
	impl plugins.Plugin
}

func (s *server) Start(ctx context.Context, req *pb.StartRequest) (*emptypb.Empty, error) {
	if err := s.impl.Start(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *server) GetSchema(ctx context.Context, _ *emptypb.Empty) (*pb.GetSchemaResponse, error) {
	fields, err := s.impl.GetSchema(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.GetSchemaResponse{Fields: fields}, nil
}

// ───────────────────────────────────────────────────────────────────────────────
//  HashiCorp go-plugin wrapper (publishes the service above)
// ───────────────────────────────────────────────────────────────────────────────

type GenericServerPlugin struct {
	plugin.NetRPCUnsupportedPlugin // fulfils legacy methods automatically
	Impl                           plugins.Plugin
}

func (p *GenericServerPlugin) GRPCServer(b *plugin.GRPCBroker, s *grpc.Server) error {
	pb.RegisterPluginServer(s, &server{impl: p.Impl})
	return nil
}

// GRPCClient is never invoked in the plugin executable.
// func (GenericServerPlugin) GRPCClient(context.Context, *plugin.GRPCBroker, *grpc.ClientConn) (interface{}, error) {
// 	return nil, nil
// }

func (GenericServerPlugin) GRPCClient(_ context.Context, _ *plugin.GRPCBroker, cc *grpc.ClientConn) (interface{}, error) {
	return &genericClient{rpc: pb.NewPluginClient(cc)}, nil
}
