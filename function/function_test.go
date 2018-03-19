package function_test

import (
	"encoding/json"
	"testing"

	"github.com/serverless/event-gateway/function"
	"github.com/stretchr/testify/assert"

	"github.com/serverless/event-gateway/providers/http"
)

func TestMarshalJSON(t *testing.T) {
	fn := &function.Function{
		Space:        "testspace",
		ID:           function.ID("testid"),
		ProviderType: http.Type,
		Provider:     &http.HTTP{URL: "http://example.com"},
	}

	data, err := json.Marshal(fn)

	assert.Nil(t, err)
	expected := []byte(`{"space":"testspace","functionId":"testid","type":"http","provider":{"url":"http://example.com"}}`)
	assert.Equal(t, expected, data)
}

func TestUnmarshalJSON(t *testing.T) {
	data := []byte(`{"space":"testspace","functionId":"testid","type":"http","provider":{"url":"http://example.com"}}`)

	fn := &function.Function{}
	err := json.Unmarshal(data, fn)

	assert.Nil(t, err)
	assert.Equal(t, function.ID("testid"), fn.ID)
	assert.Equal(t, "testspace", fn.Space)
	assert.Equal(t, http.Type, fn.ProviderType)
	assert.Equal(t, &http.HTTP{URL: "http://example.com"}, fn.Provider.(*http.HTTP))
}

func TestUnmarshalJSON_NoProvider(t *testing.T) {
	data := []byte(`{"space":"testspace","functionId":"testid"}`)

	fn := &function.Function{}
	err := json.Unmarshal(data, fn)

	assert.EqualError(t, err, "provider configuration not set")
}
