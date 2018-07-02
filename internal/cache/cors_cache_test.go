package cache

import (
	"testing"

	"github.com/serverless/event-gateway/subscription/cors"
	"github.com/stretchr/testify/assert"

	"go.uber.org/zap"
)

func TestCORSCacheModified(t *testing.T) {
	t.Run("added", func(t *testing.T) {
		ccache := newCORSCache(zap.NewNop())

		ccache.Modified("GET%2Ftest", []byte(`{
			"corsId":"GET%2Ftest",
			"space": "default",
			"method": "GET",
			"path": "/test",
			"allowedOrigins": ["*"]}`))

		value, _ := ccache.endpoints["GET"].Resolve("/test")
		config := value.(cors.CORS)
		assert.Equal(t, cors.ID("GET%2Ftest"), config.ID)
		assert.Equal(t, "default", config.Space)
	})

	t.Run("wrong payload", func(t *testing.T) {
		ccache := newCORSCache(zap.NewNop())

		ccache.Modified("GET%2Ftest", []byte(`not json`))

		assert.Nil(t, ccache.endpoints["GET"])
	})

	t.Run("deleted", func(t *testing.T) {
		ccache := newCORSCache(zap.NewNop())

		ccache.Modified("GET%2Ftest", []byte(`{
		"corsId":"GET%2Ftest",
		"space": "default",
		"method": "GET",
		"path": "/test",
		"allowedOrigins": ["*"]}`))
		ccache.Modified("GET%2Ftest1", []byte(`{
		"corsId":"GET%2Ftest1",
		"space": "default",
		"method": "GET",
		"path": "/test1",
		"allowedOrigins": ["*"]}`))
		ccache.Deleted("GET%2Ftest1", []byte(`{
			"corsId":"GET%2Ftest1",
			"space": "default",
			"method": "GET",
			"path": "/test1",
			"allowedOrigins": ["*"]}`))

		value, _ := ccache.endpoints["GET"].Resolve("/test1")
		assert.Nil(t, value)
	})
}
