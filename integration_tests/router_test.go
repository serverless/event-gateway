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

	kv, shutdown, shutdownComplete := TestingEtcd()

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

	updatedChan := make(chan struct{})
	go func() {
		for {
			res := router.TargetCache.BackingFunction(endpoints.EndpointID("POST-smilez"))
			if res != nil {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		close(updatedChan)
	}()

	select {
	case <-updatedChan:
	case <-time.After(30 * time.Second):
		panic("failed to receive endpoint configuration within 30 seconds")
	}

	res := get(testRouterServer.URL + "/smilez")

	if res != expected {
		panic("returned value was not \"" + expected + "\", unexpected value: \"" + res + "\"")
	}

	close(shutdown)
	router.Drain()
	<-shutdownComplete
}

func post(url string, thing interface{}) {
	reqBytes := &bytes.Buffer{}
	json.NewEncoder(reqBytes).Encode(thing)

	resp, err := http.Post(url, "application/json", reqBytes)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	response := &functions.Function{}
	err = json.Unmarshal(body, response)
	if err != nil {
		panic(err)
	}
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
