package serve

import (
	"context"

	pb "github.com/katasec/dstream/proto"
)

// grpcServer wraps the plugin implementation and serves over gRPC
type grpcServer struct {
	pb.UnimplementedDStreamPluginServer
	impl Plugin
}

func NewGRPCServer(impl Plugin) pb.DStreamPluginServer {
	return &grpcServer{impl: impl}
}

// Start is the gRPC handler that invokes the plugin's Start method
func (s *grpcServer) Start(ctx context.Context, req *pb.StartRequest) (*pb.StartResponse, error) {
	err := s.impl.Start(ctx, req.GetConfig())
	if err != nil {
		return nil, err
	}
	return &pb.StartResponse{
		Message: "Plugin executed successfully",
	}, nil
}
