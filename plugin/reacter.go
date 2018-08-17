package plugin

import (
	"net/rpc"

	goplugin "github.com/hashicorp/go-plugin"
	eventpkg "github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/subscription"
)

// Subscription use by plugin to indicate which event it wants to react to.
type Subscription struct {
	EventType eventpkg.TypeName
	Type      subscription.Type
}

// Reacter allows reacting on subscribed event types.
type Reacter interface {
	Subscriptions() []Subscription
	React(event eventpkg.Event) error
}

// ReacterRPCPlugin is the go-plugin's Plugin implementation.
type ReacterRPCPlugin struct {
	Reacter Reacter
}

// Server hosts ReacterServer.
func (r *ReacterRPCPlugin) Server(*goplugin.MuxBroker) (interface{}, error) {
	return &ReacterServer{Reacter: r.Reacter}, nil
}

// Client provides ReacterClient client.
func (r *ReacterRPCPlugin) Client(b *goplugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ReacterClient{client: c}, nil
}

// ReacterServer is a net/rpc compatibile structure for serving a Reacter.
type ReacterServer struct {
	Reacter Reacter
}

// Subscriptions server implementation.
func (r *ReacterServer) Subscriptions(_ interface{}, resp *ReacterSubscriptionsResponse) error {
	*resp = ReacterSubscriptionsResponse{Subscriptions: r.Reacter.Subscriptions()}
	return nil
}

// React server implementation.
func (r *ReacterServer) React(args *ReacterReactArgs, resp *ReacterReactResponse) error {
	err := r.Reacter.React(args.Event)

	*resp = ReacterReactResponse{Error: goplugin.NewBasicError(err)}
	return nil
}

// ReacterClient is a RPC implementation of Reacter.
type ReacterClient struct {
	client *rpc.Client
}

// Subscriptions call plugin implementation.
func (r *ReacterClient) Subscriptions() []Subscription {
	var resp ReacterSubscriptionsResponse
	err := r.client.Call("Plugin.Subscriptions", new(interface{}), &resp)
	if err != nil {
		return []Subscription{}
	}

	return resp.Subscriptions
}

// React calls plugin implementation.
func (r *ReacterClient) React(event eventpkg.Event) error {
	args := &ReacterReactArgs{Event: event}
	var resp ReacterReactResponse
	err := r.client.Call("Plugin.React", args, &resp)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		err = resp.Error
	}
	return err
}

// ReacterSubscriptionsResponse RPC response
type ReacterSubscriptionsResponse struct {
	Subscriptions []Subscription
}

// ReacterReactArgs RPC args
type ReacterReactArgs struct {
	Event eventpkg.Event
}

// ReacterReactResponse RPC response
type ReacterReactResponse struct {
	Error *goplugin.BasicError
}
