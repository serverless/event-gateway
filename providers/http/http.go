package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/serverless/event-gateway/function"
	"go.uber.org/zap/zapcore"
	validator "gopkg.in/go-playground/validator.v9"
)

// Type of provider.
const Type = function.ProviderType("http")

func init() {
	function.RegisterProvider(Type, ProviderLoader{})
}

// HTTP function implementation
type HTTP struct {
	URL string `json:"url" validate:"required,url"`
}

// Call HTTP endpoint.
func (h HTTP) Call(payload []byte) ([]byte, error) {
	client := http.Client{
		Timeout: time.Second * 5,
	}

	resp, err := client.Post(h.URL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, &function.ErrFunctionCallFailed{Original: err}
	}
	if resp.StatusCode == http.StatusInternalServerError {
		return nil, &function.ErrFunctionError{Original: fmt.Errorf("HTTP status code: %d", http.StatusInternalServerError)}
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// MarshalLogObject is a part of zapcore.ObjectMarshaler interface.
func (h HTTP) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("type", string(Type))
	enc.AddString("url", h.URL)
	return nil
}

// ProviderLoader implementation
type ProviderLoader struct{}

// Load decode JSON data as Config and return initialized Provider instance.
func (p ProviderLoader) Load(data []byte) (function.Provider, error) {
	provider := &HTTP{}
	err := json.Unmarshal(data, provider)
	if err != nil {
		return nil, errors.New("unable to load function provider config: " + err.Error())
	}

	validate := validator.New()
	err = validate.Struct(provider)
	if err != nil {
		return nil, &function.ErrFunctionValidation{Message: "Missing required fields for HTTP endpoint."}
	}

	return provider, nil
}
