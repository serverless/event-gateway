package router

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/julienschmidt/httprouter"
	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/httpapi"
	"github.com/serverless/event-gateway/internal/cache"
	"github.com/serverless/event-gateway/internal/embedded"
	intstore "github.com/serverless/event-gateway/internal/store"
	"github.com/serverless/event-gateway/internal/sync"
	"github.com/serverless/event-gateway/kv"
	"github.com/serverless/event-gateway/plugin"
	"github.com/serverless/event-gateway/subscription"
	"github.com/serverless/libkv"
	"github.com/serverless/libkv/store"
	etcd "github.com/serverless/libkv/store/etcd/v3"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	etcd.Register()
}

func TestIntegration_AsyncSubscription(t *testing.T) {
	logCfg := zap.NewDevelopmentConfig()
	logCfg.DisableStacktrace = true
	log, _ := logCfg.Build()
	kv, shutdownGuard := newTestEtcd()

	testAPIServer := newConfigAPIServer(kv, log)
	defer testAPIServer.Close()
	router, testRouterServer := newTestRouterServer(kv, log)
	defer testRouterServer.Close()
	router.StartWorkers()

	expected := "😸"

	// register subscriber function
	smileyReceived := make(chan struct{})
	testSubscriberServer := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			reqBuf, _ := ioutil.ReadAll(r.Body)

			var event event.Event
			err := json.Unmarshal(reqBuf, &event)
			if err != nil {
				panic(err)
			}
			decoded, _ := base64.StdEncoding.DecodeString(event.Data.(string))

			if string(decoded) == expected {
				close(smileyReceived)
			} else {
				log.Error("received non-smiley!", zap.String("value", fmt.Sprintf("%+v", reqBuf)))
			}
		}))
	defer testSubscriberServer.Close()

	subscriberFnID := function.ID("smileysubscriber")
	post(testAPIServer.URL+"/v1/functions",
		function.Function{
			ID: subscriberFnID,
			Provider: &function.Provider{
				Type: function.HTTPEndpoint,
				URL:  testSubscriberServer.URL,
			},
		})
	wait(router.WaitForFunction(subscriberFnID), "timed out waiting for function to be configured!")

	// set up pub/sub
	eventType := "smileys"

	post(testAPIServer.URL+"/v1/subscriptions", subscription.Subscription{
		FunctionID: subscriberFnID,
		Event:      event.Type(eventType),
		Path:       "/",
	})
	wait(router.WaitForSubscriber("/", event.Type(eventType)), "timed out waiting for subscriber to be configured!")

	emit(testRouterServer.URL, eventType, []byte(expected))
	wait(smileyReceived, "timed out waiting to receive pub/sub event in subscriber!")

	router.Drain()
	shutdownGuard.ShutdownAndWait()
}

func TestIntegration_HTTPSubscription(t *testing.T) {
	logCfg := zap.NewDevelopmentConfig()
	logCfg.DisableStacktrace = true
	log, _ := logCfg.Build()

	kv, shutdownGuard := newTestEtcd()

	testAPIServer := newConfigAPIServer(kv, log)
	defer testAPIServer.Close()

	router, testRouterServer := newTestRouterServer(kv, log)
	defer testRouterServer.Close()

	testTargetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"statusCode":201,"headers":{"content-type":"text/html"},"body":"<head></head>"}`)
	}))
	defer testTargetServer.Close()

	functionID := function.ID("httpresponse")
	post(testAPIServer.URL+"/v1/functions",
		function.Function{
			ID: functionID,
			Provider: &function.Provider{
				Type: function.HTTPEndpoint,
				URL:  testTargetServer.URL,
			},
		})
	wait(router.WaitForFunction(functionID), "timed out waiting for function to be configured!")

	post(testAPIServer.URL+"/v1/subscriptions", subscription.Subscription{
		FunctionID: function.ID("httpresponse"),
		Event:      "http",
		Method:     "GET",
		Path:       "/httpresponse",
	})
	wait(router.WaitForEndpoint("GET", "/httpresponse"), "timed out waiting for endpoint to be configured!")

	statusCode, headers, body := get(testRouterServer.URL + "/httpresponse")
	assert.Equal(t, statusCode, 201)
	assert.Equal(t, headers.Get("content-type"), "text/html")
	assert.Equal(t, body, "<head></head>")

	router.Drain()
	shutdownGuard.ShutdownAndWait()
}

func wait(ch <-chan struct{}, errMsg string) {
	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		panic(errMsg)
	}
}

func emit(url, eventType string, body []byte) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}

	req.Header.Add("event", eventType)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}

func post(url string, payload interface{}) ([]byte, error) {
	reqBytes := &bytes.Buffer{}
	json.NewEncoder(reqBytes).Encode(payload)

	resp, err := http.Post(url, "application/json", reqBytes)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func get(url string) (int, http.Header, string) {
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		panic(err)
	}

	return res.StatusCode, res.Header, string(body)
}

func newTestRouterServer(kvstore store.Store, log *zap.Logger) (*Router, *httptest.Server) {
	targetCache := cache.NewTarget("/serverless-event-gateway", kvstore, log)

	router := New(10, 10, targetCache, plugin.NewManager([]string{}, log), log)
	return router, httptest.NewServer(router)
}

// newConfigAPIServer creates test Configuration API server.
func newConfigAPIServer(kvstore store.Store, log *zap.Logger) *httptest.Server {
	apiRouter := httprouter.New()

	fnsDB := intstore.NewPrefixed("/serverless-event-gateway/functions", kvstore)
	fns := &kv.Functions{
		DB:  fnsDB,
		Log: log,
	}

	subs := &kv.Subscriptions{
		SubscriptionsDB: intstore.NewPrefixed("/serverless-event-gateway/subscriptions", kvstore),
		EndpointsDB:     intstore.NewPrefixed("/serverless-event-gateway/endpoints", kvstore),
		FunctionsDB:     fnsDB,
		Log:             log,
	}

	ha := &httpapi.HTTPAPI{
		Functions:     fns,
		Subscriptions: subs,
	}
	ha.RegisterRoutes(apiRouter)

	return httptest.NewServer(apiRouter)
}

// newTestEtcd returns etcd store for testing.
func newTestEtcd() (store.Store, *sync.ShutdownGuard) {
	shutdownGuard := sync.NewShutdownGuard()

	wd, err := os.Getwd()
	if err != nil {
		shutdownGuard.ShutdownAndWait()
		panic(err)
	}

	peerPort := newPort()
	peerAddr := "http://localhost:" + strconv.Itoa(peerPort)

	etcdDir := "testing.etcd"
	dataDir := wd + "/" + etcdDir + "." + strconv.Itoa(peerPort)

	cliPort := newPort()
	cliKvAddr := kvAddr(cliPort)
	cliAddr := "http://" + cliKvAddr

	embedded.EmbedEtcd(dataDir, peerAddr, cliAddr, shutdownGuard)

	cli, err := libkv.NewStore(
		store.ETCDV3,
		[]string{cliKvAddr},
		&store.Config{
			ConnectionTimeout: 10 * time.Second,
		},
	)
	if err != nil {
		shutdownGuard.ShutdownAndWait()
		panic(err)
	}

	go func() {
		shutdownGuard.Add(1)
		<-shutdownGuard.ShuttingDown
		err := os.RemoveAll(dataDir)
		shutdownGuard.Done()
		if err != nil {
			panic(err)
		}
	}()

	return cli, shutdownGuard
}

func kvAddr(port int) string {
	return "127.0.0.1:" + strconv.Itoa(port)
}

var (
	testPortAllocator = uint32(3370)
)

func newPort() int {
	return int(atomic.AddUint32(&testPortAllocator, 1))
}
