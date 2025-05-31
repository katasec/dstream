package serve

import (
	"context"

	pb "github.com/katasec/dstream/proto"
	"google.golang.org/grpc"
)

// GRPCClient implements the Plugin interface by calling the remote gRPC plugin
type GRPCClient struct {
	client pb.PluginClient
}

func NewGRPCClient(cc *grpc.ClientConn) *GRPCClient {
	return &GRPCClient{
		client: pb.NewPluginClient(cc),
	}
}

// Start calls the remote plugin's Start method over gRPC
func (g *GRPCClient) Start(ctx context.Context, cfg map[string]string) error {
	_, err := g.client.Start(ctx, &pb.StartRequest{
		Config: cfg,
	})
	return err
}

// GetSchema calls the remote plugin's GetSchema method over gRPC
func (g *GRPCClient) GetSchema(ctx context.Context) ([]*pb.FieldSchema, error) {
	resp, err := g.client.GetSchema(ctx, &pb.GetSchemaRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Fields, nil
}
