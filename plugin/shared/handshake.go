// Package shared contains shared data between the host and plugins.
package shared

import "github.com/hashicorp/go-plugin"

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	// The ProtocolVersion is the version that must match between EG core and EG plugins. This should be bumped whenever
	// a change happens in one or the other that makes it so that they can't safely communicate.
	ProtocolVersion: 1,
	// The magic cookie values should NEVER be changed.
	MagicCookieKey:   "EVENT_GATEWAY_MAGIC_COOKIE",
	MagicCookieValue: "0329c93c-a64c-4eb5-bf72-63172430d433",
}
