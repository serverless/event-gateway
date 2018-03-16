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
	Space string `validate:"required,min=3,space"`
	ID    ID     `validate:"required,functionid"`
	ProviderType
	Provider `validate:"-"`
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
	enc.AddObject("provider", f.Provider)

	return nil
}

// functionJSON is an internal struct used for JSON (un)marshaling
type functionJSON struct {
	Space string `json:"space"`
	ID    ID     `json:"functionId"`

	ProviderType   `json:"type"`
	ProviderConfig *json.RawMessage `json:"provider"`
}

// MarshalJSON marshals provides as config and returns JSON representation of the function.
func (f *Function) MarshalJSON() ([]byte, error) {
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
	rawFunction := functionJSON{}
	if err := json.Unmarshal(data, &rawFunction); err != nil {
		return err
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
