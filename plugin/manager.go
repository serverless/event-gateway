package plugin

import (
	"encoding/gob"
	"os/exec"

	"go.uber.org/zap"

	goplugin "github.com/hashicorp/go-plugin"
	"github.com/serverless/event-gateway/event"
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
	Path          string
	Client        *goplugin.Client
	Reacter       Reacter
	Subscriptions []Subscription
}

// Manager handles lifecycle of plugin management.
type Manager struct {
	Plugins []*Plugin
	Log     *zap.Logger
}

// NewManager creates new Manager.
func NewManager(paths []string, log *zap.Logger) *Manager {
	plugins := []*Plugin{}
	logger := Hclog2ZapLogger{log}
	for _, path := range paths {
		client := goplugin.NewClient(&goplugin.ClientConfig{
			HandshakeConfig: handshakeConfig,
			Plugins:         pluginMap,
			Cmd:             exec.Command(path),
			Logger:          logger.Named("PluginManager"),
		})

		plugins = append(plugins, &Plugin{
			Client: client,
			Path:   path,
		})
	}

	return &Manager{
		Plugins: plugins,
		Log:     log,
	}
}

// Connect connects to plugins.
func (m *Manager) Connect() error {
	for _, plugin := range m.Plugins {
		rpcClient, err := plugin.Client.Client()
		if err != nil {
			return err
		}

		// Request the plugin
		raw, err := rpcClient.Dispense("subscriber")
		if err != nil {
			return err
		}

		plugin.Reacter = raw.(*Subscriber)
		plugin.Subscriptions = plugin.Reacter.Subscriptions()
	}

	return nil
}

// Kill disconnects plugins and kill subprocesses.
func (m *Manager) Kill() {
	for _, plugin := range m.Plugins {
		plugin.Client.Kill()
	}
}

// React call all plugins' React method. It returns when the first error is returned by a plugin.
func (m *Manager) React(event *event.Event) error {
	for _, plugin := range m.Plugins {
		for _, subscription := range plugin.Subscriptions {
			if subscription.EventType == event.Type {
				err := plugin.Reacter.React(*event)
				if err != nil {
					m.Log.Debug("Plugin returned error.",
						zap.String("plugin", plugin.Path),
						zap.Error(err),
						zap.String("subscriptionType", string(subscription.Type)))
					if subscription.Type == Sync {
						return err
					}
				}
			}
		}
	}

	return nil
}

// handshakeConfig is used to just do a basic handshake between a plugin and host. If the handshake fails, a user
// friendly error is shown. This prevents users from executing bad plugins or executing a plugin directory. It is a UX
// feature, not a security feature.
var handshakeConfig = goplugin.HandshakeConfig{
	// The ProtocolVersion is the version that must match between EG core and EG plugins. This should be bumped whenever
	// a change happens in one or the other that makes it so that they can't safely communicate.
	ProtocolVersion: 1,
	// The magic cookie values should NEVER be changed.
	MagicCookieKey:   "EVENT_GATEWAY_MAGIC_COOKIE",
	MagicCookieValue: "0329c93c-a64c-4eb5-bf72-63172430d433",
}

// pluginMap is the map of plugins we can dispense.
var pluginMap = map[string]goplugin.Plugin{
	"subscriber": &SubscriberPlugin{},
}
