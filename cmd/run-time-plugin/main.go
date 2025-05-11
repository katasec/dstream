package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/katasec/dstream/pkg/plugins"
	"github.com/katasec/dstream/pkg/plugins/serve"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: serve.Handshake,
		Plugins: map[string]plugin.Plugin{
			"ingester": &serve.IngesterPlugin{},
		},
		Cmd:              exec.Command("./dstream-ingester-time"), // path to built binary
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})

	rpcClient, err := client.Client()
	if err != nil {
		log.Fatalf("failed to create RPC client: %v", err)
	}

	raw, err := rpcClient.Dispense("ingester")
	if err != nil {
		log.Fatalf("failed to dispense plugin: %v", err)
	}

	ing := raw.(plugins.Ingester)

	fmt.Println("🟢 Starting ingester...")
	err = ing.Start(ctx, func(e plugins.Event) error {
		fmt.Printf("🔔 Event: %+v\n", e)
		return nil
	})
	if err != nil {
		log.Fatalf("ingester failed: %v", err)
	}
}
