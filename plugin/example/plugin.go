package main

import (
	"encoding/gob"
	"log"

	"github.com/serverless/event-gateway/api"
	"github.com/serverless/event-gateway/plugin"
)

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register(api.SystemEventReceivedData{})
	gob.Register(api.SystemFunctionInvokingData{})
}

// Simple plugin demonstrating how to build plugins.
type Simple struct{}

// Subscriptions return list of events that plugin listens to.
func (s *Simple) Subscriptions() []plugin.Subscription {
	return []plugin.Subscription{
		plugin.Subscription{
			EventType: api.SystemEventReceivedType,
			Type:      plugin.Sync,
		},
	}
}

// React is called for every event that plugin subscribed to.
func (s *Simple) React(instance api.Event) error {
	switch instance.Type {
	case api.SystemEventReceivedType:
		received := instance.Data.(api.SystemEventReceivedData)
		log.Printf("received gateway.received.event for event: %q", received.Event.Type)
		break
	}

	return nil
}
