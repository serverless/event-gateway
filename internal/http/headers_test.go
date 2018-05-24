package http_test

import (
	"net/http"
	"testing"

	ihttp "github.com/serverless/event-gateway/internal/http"
	"github.com/stretchr/testify/assert"
)

func TestFlattenHeader(t *testing.T) {
	for _, testCase := range flattenHeaderTests {
		assert.Equal(t, testCase.result, ihttp.FlattenHeader(testCase.header))
	}
}

var flattenHeaderTests = []struct {
	header http.Header
	result map[string]string
}{
	{
		map[string][]string{"Custom-Header": []string{"value"}},
		map[string]string{"Custom-Header": "value"},
	},
	{
		map[string][]string{"Custom-Header": []string{"value1", "value2"}},
		map[string]string{"Custom-Header": "value1, value2"},
	},
}
