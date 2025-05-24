package serve

import (
	"context"

	pb "github.com/katasec/dstream/proto"
)

// grpcServer wraps a user-defined Plugin and exposes it over gRPC.
type grpcServer struct {
	pb.UnimplementedDStreamPluginServer
	impl Plugin
}

// NewGRPCServer returns a gRPC server implementation of DStreamPlugin
func NewGRPCServer(impl Plugin) pb.DStreamPluginServer {
	return &grpcServer{impl: impl}
}

// Start receives a config blob and calls the plugin's Start method
func (s *grpcServer) Start(ctx context.Context, req *pb.StartRequest) (*pb.StartResponse, error) {
	err := s.impl.Start(ctx, req.GetConfig())
	if err != nil {
		return nil, err
	}
	return &pb.StartResponse{Message: "Plugin executed successfully"}, nil
}
