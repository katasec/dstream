package serve

import (
	"context"

	"github.com/hashicorp/go-plugin"
	pb "github.com/katasec/dstream/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ───────────────────────────────────────────────────────────────────────────────
//  HashiCorp go-plugin wrapper (lives in the *CLI* process)
// ───────────────────────────────────────────────────────────────────────────────

// GenericPlugin satisfies plugin.Plugin so the executor can establish a GRPC
// connection to the external binary. Net-RPC stubs are supplied by the embedded
// helper.
type GenericPlugin struct {
	plugin.NetRPCUnsupportedPlugin
}

// GRPCClient returns a client-side implementation of plugins.Plugin.
func (GenericPlugin) GRPCClient(_ context.Context, _ *plugin.GRPCBroker, cc *grpc.ClientConn) (interface{}, error) {
	return &genericClient{rpc: pb.NewPluginClient(cc)}, nil
}

// GRPCServer is not used on the CLI side.
func (GenericPlugin) GRPCServer(*plugin.GRPCBroker, *plugin.GRPCServer) error { return nil }

// ───────────────────────────────────────────────────────────────────────────────
//  plugins.Plugin façade used by the executor
// ───────────────────────────────────────────────────────────────────────────────

type genericClient struct{ rpc pb.PluginClient }

func (c *genericClient) Start(ctx context.Context, req *pb.StartRequest) error {
	_, err := c.rpc.Start(ctx, req)
	return err
}

func (c *genericClient) GetSchema(ctx context.Context) ([]*pb.FieldSchema, error) {
	resp, err := c.rpc.GetSchema(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	return resp.Fields, nil
}
