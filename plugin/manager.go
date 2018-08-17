package plugin

import (
	"encoding/gob"
	"os/exec"

	"go.uber.org/zap"

	goplugin "github.com/hashicorp/go-plugin"
	"github.com/serverless/event-gateway/event"
	eventpkg "github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/plugin/shared"
	"github.com/serverless/event-gateway/subscription"
)

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register(event.SystemEventReceivedData{})
	gob.Register(event.SystemFunctionInvokingData{})
	gob.Register(event.SystemFunctionInvokedData{})
	gob.Register(event.SystemFunctionInvocationFailedData{})
}

// Plugin is a generic struct for storing info about a plugin.
type Plugin struct {
	Path    string
	Reacter Reacter
	Queue   Queue
}

// Manager handles lifecycle of plugin management.
type Manager struct {
	Reacters []*Plugin
	Queue    *Plugin
	Log      *zap.Logger
}

// NewManager creates new Manager.
func NewManager(paths []string, log *zap.Logger) (*Manager, error) {
	reacters := []*Plugin{}
	queue := &Plugin{}
	logger := Hclog2ZapLogger{log}

	for _, path := range paths {
		client := goplugin.NewClient(&goplugin.ClientConfig{
			HandshakeConfig: shared.Handshake,
			Plugins:         pluginMap,
			Cmd:             exec.Command(path),
			Logger:          logger.Named("PluginManager"),
			Managed:         true,
		})

		rpcClient, err := client.Client()
		if err != nil {
			return nil, err
		}

		// Request the plugin
		raw, _ := rpcClient.Dispense("reacter")
		if raw == nil {
			raw, _ = rpcClient.Dispense("queue")
		}

		switch raw.(type) {
		case *QueueClient:
			queue.Path = path
			queue.Queue = raw.(*QueueClient)
		case *ReacterClient:
			reacters = append(reacters, &Plugin{
				Path:    path,
				Reacter: raw.(*ReacterClient),
			})
		default:
			panic("plugin unrecognized")
		}
	}

	return &Manager{
		Reacters: reacters,
		Queue:    queue,
		Log:      log,
	}, nil
}

// Kill disconnects plugins and kill subprocesses.
func (m *Manager) Kill() {
	goplugin.CleanupClients()
}

// React call all plugins' React method. It returns when the first error is returned by a plugin.
func (m *Manager) React(event *event.Event) error {
	for _, plugin := range m.Reacters {
		for _, sub := range plugin.Reacter.Subscriptions() {
			if sub.EventType == event.EventType {
				err := plugin.Reacter.React(*event)
				if err != nil {
					m.Log.Debug("Plugin returned error.",
						zap.String("plugin", plugin.Path),
						zap.Error(err),
						zap.String("subscriptionType", string(sub.Type)))
					if sub.Type == subscription.TypeSync {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (m *Manager) Push(subscriptionID subscription.ID, event eventpkg.Event) error {
	return m.Queue.Queue.Push(subscriptionID, event)
}
func (m *Manager) Pull() (*eventpkg.Event, error) {
	return nil, nil
}
func (m *Manager) MarkAsDelivered(subscriptionID subscription.ID, eventID string) error {
	return nil
}

// pluginMap is the map of plugins we can dispense.
var pluginMap = map[string]goplugin.Plugin{
	"reacter": &ReacterRPCPlugin{},
	"queue":   &QueueRPCPlugin{},
}
