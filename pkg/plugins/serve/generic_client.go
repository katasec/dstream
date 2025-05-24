package serve

import (
	"context"

	pb "github.com/katasec/dstream/proto"
	"google.golang.org/grpc"
)

// GRPCClient implements the Plugin interface by calling the remote gRPC plugin
type GRPCClient struct {
	client pb.DStreamPluginClient
}

func NewGRPCClient(cc *grpc.ClientConn) *GRPCClient {
	return &GRPCClient{
		client: pb.NewDStreamPluginClient(cc),
	}
}

// Start calls the remote plugin's Start method over gRPC
func (g *GRPCClient) Start(ctx context.Context, rawConfig []byte) error {
	_, err := g.client.Start(ctx, &pb.StartRequest{
		Config: rawConfig,
	})
	return err
}
