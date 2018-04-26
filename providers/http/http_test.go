package http_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	httpprovider "github.com/serverless/event-gateway/providers/http"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	config := []byte(`{"url":""}`)
	loader := httpprovider.ProviderLoader{}

	provider, err := loader.Load(config)

	assert.Nil(t, provider)
	assert.EqualError(t, err, "missing required fields for HTTP endpoint")
}

func TestCall(t *testing.T) {
    var contentType string
	echo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	    contentType = r.Header.Get("Content-Type")
		payload, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		fmt.Fprint(w, string(payload))
	}))
	provider := httpprovider.HTTP{
		URL: echo.URL,
	}

	resp, err := provider.Call([]byte("hello"))

	assert.Nil(t, err)
	assert.Equal(t, "application/cloudevents+json", contentType)
	assert.Equal(t, "hello", string(resp))
}

func TestCall_InternalError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	provider := httpprovider.HTTP{
		URL: ts.URL,
	}

	_, err := provider.Call([]byte("hello"))

	assert.EqualError(t, err, "Function call failed because of runtime error. Error: HTTP status code: 500")
}
