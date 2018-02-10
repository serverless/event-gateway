package plugin

import (
	"net/rpc"

	goplugin "github.com/hashicorp/go-plugin"
	eventpkg "github.com/serverless/event-gateway/event"
)

// Reacter provides subscriptions for events that can react to.
type Reacter interface {
	Subscriptions() []Subscription
	React(event eventpkg.Event) error
}

// Type of a subscription.
type Type string

const (
	// Async subscription type. Plugin host will not block on it.
	Async Type = "async"
	// Sync subscription type. Plugin host will use the response from the plugin before proceeding.
	Sync Type = "sync"
)

// Subscription use by plugin to indicate which event it want to react to.
type Subscription struct {
	EventType eventpkg.Type
	Type      Type
}

// Subscriber is a RPC implementation of Reacter.
type Subscriber struct {
	client *rpc.Client
}

// Subscriptions call plugin implementation.
func (s *Subscriber) Subscriptions() []Subscription {
	var resp SubscriberSubscriptionsResponse
	err := s.client.Call("Plugin.Subscriptions", new(interface{}), &resp)
	if err != nil {
		return []Subscription{}
	}

	return resp.Subscriptions
}

// SubscriberSubscriptionsResponse RPC response
type SubscriberSubscriptionsResponse struct {
	Subscriptions []Subscription
}

// React calls plugin implementation.
func (s *Subscriber) React(event eventpkg.Event) error {
	args := &SubscriberReactArgs{Event: event}
	var resp SubscriberReactResponse
	err := s.client.Call("Plugin.React", args, &resp)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		err = resp.Error
	}

	return err
}

// SubscriberReactArgs RPC args
type SubscriberReactArgs struct {
	Event eventpkg.Event
}

// SubscriberReactResponse RPC response
type SubscriberReactResponse struct {
	Error *goplugin.BasicError
}

// SubscriberServer is a net/rpc compatibile structure for serving a Reacter.
type SubscriberServer struct {
	Reacter Reacter
}

// Subscriptions server implementation.
func (s *SubscriberServer) Subscriptions(_ interface{}, resp *SubscriberSubscriptionsResponse) error {
	*resp = SubscriberSubscriptionsResponse{Subscriptions: s.Reacter.Subscriptions()}
	return nil
}

// React server implementation.
func (s *SubscriberServer) React(args *SubscriberReactArgs, resp *SubscriberReactResponse) error {
	err := s.Reacter.React(args.Event)

	*resp = SubscriberReactResponse{Error: goplugin.NewBasicError(err)}
	return nil
}
