package main

import (
	goplugin "github.com/hashicorp/go-plugin"
	"github.com/serverless/event-gateway/plugin"
	"github.com/serverless/event-gateway/plugin/shared"
)

func main() {
	pluginMap := map[string]goplugin.Plugin{
		"reacter": &plugin.ReacterRPCPlugin{Reacter: &Simple{}},
	}

	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins:         pluginMap,
	})
}
