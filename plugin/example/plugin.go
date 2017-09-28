package main

import (
	"errors"

	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/plugin"
)

type Simple struct{}

func (s *Simple) Subscriptions() []plugin.Subscription {
	return []plugin.Subscription{
		plugin.Subscription{
			EventType: event.Type("user.created"),
			Type:      plugin.Async,
		},
	}
}

func (s *Simple) React(event event.Event) error {
	return errors.New("error on reacting")
}
