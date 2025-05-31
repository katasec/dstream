package serve

import (
	"context"

	pb "github.com/katasec/dstream/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

// grpcServer wraps the plugin implementation and serves over gRPC
type grpcServer struct {
	pb.UnimplementedPluginServer
	impl Plugin
}

func NewGRPCServer(impl Plugin) pb.PluginServer {
	return &grpcServer{impl: impl}
}

// Start is the gRPC handler that invokes the plugin's Start method
func (s *grpcServer) Start(ctx context.Context, req *pb.StartRequest) (*emptypb.Empty, error) {
	err := s.impl.Start(ctx, req.GetConfig())
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// ✅ NEW: GetSchema implementation
func (s *grpcServer) GetSchema(ctx context.Context, req *pb.GetSchemaRequest) (*pb.GetSchemaResponse, error) {
	fields, err := s.impl.GetSchema(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.GetSchemaResponse{Fields: fields}, nil
}
