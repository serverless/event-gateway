package integrationtests

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

	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/integrationtests/stub"
	"github.com/serverless/event-gateway/metrics"
	"github.com/serverless/event-gateway/pubsub"
	"github.com/serverless/event-gateway/router"
	"github.com/serverless/event-gateway/targetcache"
)

func TestSubscriptionHTTP(t *testing.T) {
	logCfg := zap.NewDevelopmentConfig()
	logCfg.DisableStacktrace = true
	log, _ := logCfg.Build()

	kv, shutdownGuard := stub.TestEtcd()

	testAPIServer := stub.ConfigAPIServer(kv, log)
	defer testAPIServer.Close()

	router, testRouterServer := newTestRouterServer(kv, log)
	defer testRouterServer.Close()

	expected := "😸"

	testTargetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, expected)
	}))
	defer testTargetServer.Close()

	post(testAPIServer.URL+"/v1/functions",
		functions.Function{
			ID: functions.FunctionID("super smiley function"),
			Provider: &functions.Provider{
				Type: functions.HTTPEndpoint,
				URL:  testTargetServer.URL,
			},
		})

	post(testAPIServer.URL+"/v1/subscriptions", pubsub.Subscription{
		FunctionID: functions.FunctionID("super smiley function"),
		Event:      "http",
		Method:     "POST",
		Path:       "/smilez",
	})

	select {
	case <-router.WaitForEndpoint(pubsub.NewEndpointID("POST", "/smilez")):
	case <-time.After(10 * time.Second):
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

func newTestRouterServer(kv store.Store, log *zap.Logger) (*router.Router, *httptest.Server) {
	targetCache := targetcache.New("/serverless-event-gateway", kv, log, true)
	router := router.New(targetCache, metrics.DroppedPubSubEvents, log)

	return router, httptest.NewServer(router)
}
