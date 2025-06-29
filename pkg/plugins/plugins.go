package plugins

import (
	"context"

	pb "github.com/katasec/dstream/proto"
)

// Plugin is the single interface every external repo implements.
// No HashiCorp-go-plugin types leak in here.
type Plugin interface {
	// Start runs the plugin with input, output, and global configuration.
	Start(ctx context.Context, req *pb.StartRequest) error

	// GetSchema advertises configurable fields for validation/UX.
	GetSchema(ctx context.Context) ([]*pb.FieldSchema, error)
}
