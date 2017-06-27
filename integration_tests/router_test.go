package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/docker/libkv/store"
	"go.uber.org/zap"

	"github.com/serverless/event-gateway/endpoints"
	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/metrics"
	"github.com/serverless/event-gateway/router"
	"github.com/serverless/event-gateway/targetcache"
)

func newTestRouterServer(kv store.Store, log *zap.Logger) (*router.Router, *httptest.Server) {
	targetCache := targetcache.New("/serverless-gateway", kv, log, true)
	router := router.New(targetCache, metrics.DroppedPubSubEvents, log)

	return router, httptest.NewServer(router)
}

func TestFunctionDefAndCalling(t *testing.T) {
	logCfg := zap.NewDevelopmentConfig()
	logCfg.DisableStacktrace = true
	log, _ := logCfg.Build()

	kv, shutdownGuard := TestingEtcd()

	testAPIServer := newTestAPIServer(kv, log)
	defer testAPIServer.Close()

	router, testRouterServer := newTestRouterServer(kv, log)
	defer testRouterServer.Close()

	expected := "😸"

	testTargetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, expected)
	}))
	defer testTargetServer.Close()

	post(testAPIServer.URL+"/v0/gateway/api/function",
		functions.Function{
			ID: functions.FunctionID("super smiley function"),
			HTTP: &functions.HTTPProperties{
				URL: testTargetServer.URL,
			},
		})

	post(testAPIServer.URL+"/v0/gateway/api/endpoint", endpoints.Endpoint{
		FunctionID: functions.FunctionID("super smiley function"),
		Method:     "POST",
		Path:       "/smilez",
	})

	select {
	case <-router.WaitForEndpoint(endpoints.EndpointID("POST-smilez")):
	case <-time.After(5 * time.Second):
		panic("timed out waiting for endpoint to be configured!")
	}

	res := get(testRouterServer.URL + "/smilez")

	if res != expected {
		panic("returned value was not \"" + expected + "\", unexpected value: \"" + res + "\"")
	}

	router.Drain()
	shutdownGuard.ShutdownAndWait()
}

func post(url string, thing interface{}) ([]byte, error) {
	reqBytes := &bytes.Buffer{}
	json.NewEncoder(reqBytes).Encode(thing)

	resp, err := http.Post(url, "application/json", reqBytes)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func get(url string) string {
	res, err := http.Post(url, "application/json", nil)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		panic(err)
	}

	return string(body)
}
