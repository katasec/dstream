package plugins

import (
	"context"
)

type Event map[string]any

type Ingester interface {
	Start(ctx context.Context, emit func(Event) error) error
	Stop() error
}
