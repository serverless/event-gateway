package function

import (
	"encoding/json"
	"errors"

	"go.uber.org/zap/zapcore"
)

// ID uniquely identifies a function.
type ID string

// Function represents a function deployed on one of the supported providers.
type Function struct {
	Space          string           `json:"space" validate:"required,min=3,space"`
	ID             ID               `json:"functionId" validate:"required,functionid"`
	ProviderType   ProviderType     `json:"type"`
	ProviderConfig *json.RawMessage `json:"provider"`
	Provider       Provider         `json:"-" validate:"-"`

	Metadata map[string]string `json:"metadata,omitempty"`
}

// Functions is an array of functions.
type Functions []*Function

// Call tries to send a payload to a target function
func (f *Function) Call(payload []byte) ([]byte, error) {
	return f.Provider.Call(payload)
}

// MarshalLogObject is a part of zapcore.ObjectMarshaler interface
func (f Function) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("space", string(f.Space))
	enc.AddString("functionId", string(f.ID))
	enc.AddString("type", string(f.ProviderType))
	enc.AddObject("provider", f.Provider)

	return nil
}

// MarshalJSON marshals provides as config and returns JSON representation of the function.
func (f *Function) MarshalJSON() ([]byte, error) {
	// This line is needed to avoid stack overflow because of recursive MarshalJSON call
	type functionJSON Function

	config, err := json.Marshal(f.Provider)
	if err != nil {
		return nil, err
	}

	rawConfig := json.RawMessage(config)
	fn := functionJSON{
		Space:          f.Space,
		ID:             f.ID,
		ProviderType:   f.ProviderType,
		ProviderConfig: &rawConfig,
	}

	return json.Marshal(fn)
}

// UnmarshalJSON unmarshals function JSON, detects provider type and loads the provider.
func (f *Function) UnmarshalJSON(data []byte) error {
	// This line is needed to avoid stack overflow because of recursive UnmarshalJSON call
	type functionJSON Function

	rawFunction := functionJSON{}
	if err := json.Unmarshal(data, &rawFunction); err != nil {
		return err
	}

	if rawFunction.ProviderType == "" {
		return errors.New("provider configuration not set")
	}

	f.ID = rawFunction.ID
	f.Space = rawFunction.Space
	f.ProviderType = rawFunction.ProviderType

	if loader, ok := providers[rawFunction.ProviderType]; ok {
		// err includes validation errors happening on provider side
		provider, err := loader.Load(*rawFunction.ProviderConfig)
		if err != nil {
			return err
		}

		f.Provider = provider
		return nil
	}

	return errors.New("provider " + string(rawFunction.ProviderType) + " not supported")
}
