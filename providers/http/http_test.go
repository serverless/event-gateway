package http_test

import (
	"testing"

	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/providers/http"
	"github.com/stretchr/testify/assert"
)

func TestValidate_MissingURL(t *testing.T) {
	provider := http.HTTP{}

	err := provider.Validate()

	assert.Equal(t, err, &function.ErrFunctionValidation{Message: "Missing required fields for HTTP endpoint."})
}
