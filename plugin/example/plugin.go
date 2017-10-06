package main

import (
	"encoding/gob"

	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/plugin"
)

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register(event.SystemEventReceived{})
	gob.Register(event.SystemFunctionInvoking{})
}

// Simple plugin demonstrating how to build plugins.
type Simple struct{}

// Subscriptions return list of events that plugin listens to.
func (s *Simple) Subscriptions() []plugin.Subscription {
	return []plugin.Subscription{
		plugin.Subscription{
			EventType: event.SystemEventReceivedType,
			Type:      plugin.Sync,
		},
		plugin.Subscription{
			EventType: event.SystemFunctionInvokingType,
			Type:      plugin.Sync,
		},
	}
}

// React is called for every event that plugin subscribed to.
func (s *Simple) React(event event.Event) error {
	return nil
}
