package plugins

import (
	"context"

	pb "github.com/katasec/dstream/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// Plugin is the single interface every external repo implements.
// No HashiCorp-go-plugin types leak in here.
type Plugin interface {
	// Start runs the plugin with its hierarchical config block.
	Start(ctx context.Context, cfg *structpb.Struct) error

	// GetSchema advertises configurable fields for validation/UX.
	GetSchema(ctx context.Context) ([]*pb.FieldSchema, error)
}
