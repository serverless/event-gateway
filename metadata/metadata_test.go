package metadata_test

import (
	"testing"

	"github.com/serverless/event-gateway/metadata"
	"github.com/stretchr/testify/assert"
)

func TestCheck(t *testing.T) {
	t.Run("return false if metadata value doesn't match", func(t *testing.T) {
		md := metadata.Metadata{"testkey1": "testvalue1"}

		assert.False(t, md.Check(metadata.Filter{Key: "testkey1", Value: "nottestvalue1"}))
	})

	t.Run("return false if filter doesn't apply", func(t *testing.T) {
		md := metadata.Metadata{"testkey1": "testvalue1"}

		assert.False(t, md.Check(
			metadata.Filter{Key: "testkey3", Value: "testvalue3"},
		))
	})

	t.Run("return true if filter does applies", func(t *testing.T) {
		md := metadata.Metadata{"testkey1": "testvalue1", "testkey2": "testvalue2"}

		assert.True(t, md.Check(
			metadata.Filter{Key: "testkey1", Value: "testvalue1"},
			metadata.Filter{Key: "testkey2", Value: "testvalue2"},
		))
	})
}
