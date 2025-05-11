package ingester

import (
	"context"
	"encoding/json"

	plugins "github.com/katasec/dstream/pkg/plugins"
	pb "github.com/katasec/dstream/proto"
)

// GRPCClient implements plugins.Ingester and wraps the gRPC stub
type GRPCClient struct {
	client pb.IngesterPluginClient
}

func NewGRPCClient(c pb.IngesterPluginClient) *GRPCClient {
	return &GRPCClient{client: c}
}

func (c *GRPCClient) Start(ctx context.Context, emit func(plugins.Event) error) error {
	// Send basic config if needed; for now, just empty JSON
	req := &pb.StreamRequest{ConfigJson: "{}"}

	stream, err := c.client.Start(ctx, req)
	if err != nil {
		return err
	}

	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}

		var event plugins.Event
		if err := json.Unmarshal([]byte(msg.JsonPayload), &event); err != nil {
			return err
		}

		if err := emit(event); err != nil {
			return err
		}
	}
}

func (c *GRPCClient) Stop() error {
	// Not supported in this simple example
	return nil
}
