package function

import "go.uber.org/zap/zapcore"

type ProviderType string

// Provider is an interface for function provider.
type Provider interface {
	Call(payload []byte) ([]byte, error)
	MarshalLogObject(enc zapcore.ObjectEncoder) error
}

// ProviderLoader returns Provider instance based on JSON config blob.
type ProviderLoader interface {
	Load(config []byte) (Provider, error)
}

// Registered providers loaders.
var providers = make(map[ProviderType]ProviderLoader)

// RegisterProvider registers provider loader by provider type.
func RegisterProvider(providerType ProviderType, loader ProviderLoader) {
	providers[providerType] = loader
}
