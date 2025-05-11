package main

import (
	"context"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/katasec/dstream/pkg/plugins"
	"github.com/katasec/dstream/pkg/plugins/serve"
)

type TimeIngester struct{}

func (t *TimeIngester) Start(ctx context.Context, emit func(plugins.Event) error) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case now := <-ticker.C:
			event := plugins.Event{"now": now.Format(time.RFC3339)}
			if err := emit(event); err != nil {
				return err
			}
		}
	}
}

func (t *TimeIngester) Stop() error {
	return nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: serve.Handshake,
		Plugins: map[string]plugin.Plugin{
			"ingester": &serve.IngesterPlugin{
				Impl: &TimeIngester{},
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
