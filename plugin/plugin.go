package plugin

import (
	"net/rpc"

	goplugin "github.com/hashicorp/go-plugin"
)

// SubscriberPlugin is the plugin.Plugin implementation.
type SubscriberPlugin struct {
	Reacter Reacter
}

// Server hosts SubscriberServer.
func (s *SubscriberPlugin) Server(*goplugin.MuxBroker) (interface{}, error) {
	return &SubscriberServer{Reacter: s.Reacter}, nil
}

// Client provides Subscriber client.
func (s *SubscriberPlugin) Client(b *goplugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &Subscriber{client: c}, nil
}
