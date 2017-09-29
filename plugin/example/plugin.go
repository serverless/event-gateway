package main

import (
	"errors"

	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/plugin"
)

// Simple plugin demonstrating how to build plugins.
type Simple struct{}

// Subscriptions return list of events that plugin listens to.
func (s *Simple) Subscriptions() []plugin.Subscription {
	return []plugin.Subscription{
		plugin.Subscription{
			EventType: event.Type("gateway.event.received"),
			Type:      plugin.Async,
		},
	}
}

// React is called for every event that plugin subscribed to.
func (s *Simple) React(event event.Event) error {
	return errors.New("error on reacting")
}
