package http_test

import (
	"testing"

	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/providers/http"
	"github.com/stretchr/testify/assert"
)

func TestLoad_MissingURL(t *testing.T) {
	loader := &http.ProviderLoader{}

	_, err := loader.Load([]byte(`{}`))

	assert.Equal(t, err, &function.ErrFunctionValidation{Message: "Missing required fields for HTTP endpoint."})
}
