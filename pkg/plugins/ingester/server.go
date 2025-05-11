package ingester

import (
	"encoding/json"

	plugins "github.com/katasec/dstream/pkg/plugins"
	pb "github.com/katasec/dstream/proto"
	"google.golang.org/grpc"
)

// GRPCServer wraps a plugins.Ingester and serves it over gRPC
type GRPCServer struct {
	Impl plugins.Ingester
	pb.UnimplementedIngesterPluginServer
}

// RegisterGRPC registers the gRPC server with HashiCorp plugin system
func (s *GRPCServer) RegisterGRPC(grpcServer *grpc.Server) error {
	pb.RegisterIngesterPluginServer(grpcServer, s)
	return nil
}

// Start handles the gRPC call from the CLI and streams events
func (s *GRPCServer) Start(req *pb.StreamRequest, stream pb.IngesterPlugin_StartServer) error {
	ctx := stream.Context()

	// parse config if needed (optional enhancement)
	config := req.GetConfigJson()

	_ = config // Currently unused – you can forward it if needed

	return s.Impl.Start(ctx, func(e plugins.Event) error {
		// convert Event (map[string]any) to JSON
		data, err := json.Marshal(e)
		if err != nil {
			return err
		}
		return stream.Send(&pb.Event{JsonPayload: string(data)})
	})
}
