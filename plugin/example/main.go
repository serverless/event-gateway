package main

import (
	goplugin "github.com/hashicorp/go-plugin"
	"github.com/serverless/event-gateway/plugin"
)

func main() {
	pluginMap := map[string]goplugin.Plugin{
		"subscriber": &plugin.SubscriberPlugin{Reacter: &Simple{}},
	}

	handshakeConfig := goplugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "EVENT_GATEWAY_MAGIC_COOKIE",
		MagicCookieValue: "0329c93c-a64c-4eb5-bf72-63172430d433",
	}

	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
