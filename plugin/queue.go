package plugin

import (
	"net/rpc"

	goplugin "github.com/hashicorp/go-plugin"
	eventpkg "github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/subscription"
)

type Queue interface {
	Push(subscriptionID subscription.ID, event eventpkg.Event) error
	Pull() (*eventpkg.Event, error)
	MarkAsDelivered(subscriptionID subscription.ID, eventID string) error
}

// QueueRPCPlugin is the go-plugin's Plugin implementation.
type QueueRPCPlugin struct {
	Queue
}

// Server hosts QueueServer.
func (q *QueueRPCPlugin) Server(*goplugin.MuxBroker) (interface{}, error) {
	return &QueueServer{Queue: q.Queue}, nil
}

// Client provides QueueClient client.
func (q *QueueRPCPlugin) Client(b *goplugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &QueueClient{client: c}, nil
}

// QueueServer is a net/rpc compatibile structure for serving a Queue.
type QueueServer struct {
	Queue
}

// Push server implementation.
func (q *QueueServer) Push(args *QueuePushArgs, resp *QueuePushResponse) error {
	err := q.Queue.Push(args.SubscriptionID, args.Event)

	*resp = QueuePushResponse{Error: goplugin.NewBasicError(err)}
	return nil
}

// Pull server implementation.
func (q *QueueServer) Pull(_ interface{}, resp *QueuePullResponse) error {
	event, err := q.Queue.Pull()

	*resp = QueuePullResponse{Event: *event, Error: goplugin.NewBasicError(err)}
	return nil
}

// MarkAsDelivered server implementation.
func (q *QueueServer) MarkAsDelivered(args *QueueMarkAsDeliveredArgs, resp *QueueMarkAsDeliveredResponse) error {
	err := q.Queue.MarkAsDelivered(args.SubscriptionID, args.EventID)

	*resp = QueueMarkAsDeliveredResponse{Error: goplugin.NewBasicError(err)}
	return nil
}

// QueueClient is a RPC implementation of Queue.
type QueueClient struct {
	client *rpc.Client
}

func (q *QueueClient) Push(subscriptionID subscription.ID, event eventpkg.Event) error {
	args := &QueuePushArgs{SubscriptionID: subscriptionID, Event: event}
	var resp QueuePushResponse
	err := q.client.Call("Plugin.Push", args, &resp)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		err = resp.Error
	}
	return err
}

func (q *QueueClient) Pull() (*eventpkg.Event, error) {
	var resp QueuePullResponse
	err := q.client.Call("Plugin.Pull", new(interface{}), &resp)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		err = resp.Error
	}
	return &resp.Event, err
}

func (q *QueueClient) MarkAsDelivered(subscriptionID subscription.ID, eventID string) error {
	args := &QueueMarkAsDeliveredArgs{SubscriptionID: subscriptionID, EventID: eventID}
	var resp QueueMarkAsDeliveredResponse
	err := q.client.Call("Plugin.MarkAsDelivered", args, &resp)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		err = resp.Error
	}
	return err
}

type QueuePushArgs struct {
	SubscriptionID subscription.ID
	Event          eventpkg.Event
}

type QueuePushResponse struct {
	Error *goplugin.BasicError
}

type QueuePullResponse struct {
	Event eventpkg.Event
	Error *goplugin.BasicError
}

type QueueMarkAsDeliveredArgs struct {
	SubscriptionID subscription.ID
	EventID        string
}

type QueueMarkAsDeliveredResponse struct {
	Error *goplugin.BasicError
}
