// +build integration

package tests

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
	eventgateway "github.com/serverless/event-gateway/libkv"
	"github.com/serverless/event-gateway/plugin"
	httpprovider "github.com/serverless/event-gateway/providers/http"
	"github.com/serverless/event-gateway/router"
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
	instance, testRouterServer := newTestRouterServer(kv, log)
	defer testRouterServer.Close()
	instance.StartWorkers()

	expected := "😸"
	eventType := "smileys"

	// register event type
	eventTypeName := event.TypeName(eventType)
	postEventType(testAPIServer.URL+"/v1/spaces/default/eventtypes", &event.Type{Name: eventTypeName})
	wait(instance.WaitForEventType("default", eventTypeName), "timed out waiting for event type to be configured!")

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
	postFunction(testAPIServer.URL+"/v1/spaces/default/functions",
		&function.Function{
			ID:           subscriberFnID,
			ProviderType: httpprovider.Type,
			Provider: &httpprovider.HTTP{
				URL: testSubscriberServer.URL,
			},
		})
	wait(instance.WaitForFunction("default", subscriberFnID), "timed out waiting for function to be configured!")

	// set up pub/sub
	postSubscription(testAPIServer.URL+"/v1/spaces/default/subscriptions", &subscription.Subscription{
		FunctionID: subscriberFnID,
		Type:       subscription.TypeAsync,
		EventType:  event.TypeName(eventType),
		Path:       "/",
	})
	wait(instance.WaitForAsyncSubscriber(http.MethodPost, "/", event.TypeName(eventType)), "timed out waiting for subscriber to be configured!")

	// emit event
	emit(testRouterServer.URL, eventType, []byte(expected))
	wait(smileyReceived, "timed out waiting to receive pub/sub event in subscriber!")

	instance.Drain()
	shutdownGuard.ShutdownAndWait()
}

func TestIntegration_SyncSubscription(t *testing.T) {
	logCfg := zap.NewDevelopmentConfig()
	logCfg.DisableStacktrace = true
	log, _ := logCfg.Build()
	kv, shutdownGuard := newTestEtcd()

	testAPIServer := newConfigAPIServer(kv, log)
	defer testAPIServer.Close()
	instance, testRouterServer := newTestRouterServer(kv, log)
	defer testRouterServer.Close()

	testTargetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"statusCode":201,"headers":{"content-type":"text/html"},"body":"<head></head>"}`)
	}))
	defer testTargetServer.Close()

	// register event type
	postEventType(testAPIServer.URL+"/v1/spaces/default/eventtypes", &event.Type{Name: event.TypeHTTPRequest})
	wait(instance.WaitForEventType("default", event.TypeHTTPRequest), "timed out waiting for event type to be configured!")

	// register subscriber function
	functionID := function.ID("httpresponse")
	postFunction(testAPIServer.URL+"/v1/spaces/default/functions",
		&function.Function{
			ID:           functionID,
			ProviderType: httpprovider.Type,
			Provider: &httpprovider.HTTP{
				URL: testTargetServer.URL,
			},
		})
	wait(instance.WaitForFunction("default", functionID), "timed out waiting for function to be configured!")

	// set up pub/sub
	postSubscription(testAPIServer.URL+"/v1/spaces/default/subscriptions", &subscription.Subscription{
		FunctionID: function.ID("httpresponse"),
		Type:       subscription.TypeSync,
		EventType:  event.TypeHTTPRequest,
		Method:     "GET",
		Path:       "/httpresponse",
	})
	wait(instance.WaitForSyncSubscriber(http.MethodGet, "/httpresponse", event.TypeHTTPRequest), "timed out waiting for endpoint to be configured!")

	// request function
	statusCode, headers, body := get(testRouterServer.URL + "/httpresponse")
	assert.Equal(t, 201, statusCode)
	assert.Equal(t, "text/html", headers.Get("content-type"))
	assert.Equal(t, "<head></head>", body)

	instance.Drain()
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

func postEventType(url string, eventType *event.Type) ([]byte, error) {
	reqBytes := &bytes.Buffer{}
	json.NewEncoder(reqBytes).Encode(eventType)
	return post(url, reqBytes)
}

func postFunction(url string, fn *function.Function) ([]byte, error) {
	reqBytes := &bytes.Buffer{}
	json.NewEncoder(reqBytes).Encode(fn)
	return post(url, reqBytes)
}

func postSubscription(url string, sub *subscription.Subscription) ([]byte, error) {
	reqBytes := &bytes.Buffer{}
	json.NewEncoder(reqBytes).Encode(sub)
	return post(url, reqBytes)
}

func post(url string, payload *bytes.Buffer) ([]byte, error) {
	resp, err := http.Post(url, "application/json", payload)
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

func newTestRouterServer(kvstore store.Store, log *zap.Logger) (*router.Router, *httptest.Server) {
	targetCache := cache.NewTarget("/serverless-event-gateway", kvstore, log)
	pluginManager, _ := plugin.NewManager([]string{}, log)
	instance := router.New(10, 10, targetCache, pluginManager, log)
	return instance, httptest.NewServer(instance)
}

// newConfigAPIServer creates test Configuration API server.
func newConfigAPIServer(kvstore store.Store, log *zap.Logger) *httptest.Server {
	apiRouter := httprouter.New()

	service := &eventgateway.Service{
		EventTypeStore:    intstore.NewPrefixed("/serverless-event-gateway/eventtypes", kvstore),
		FunctionStore:     intstore.NewPrefixed("/serverless-event-gateway/functions", kvstore),
		SubscriptionStore: intstore.NewPrefixed("/serverless-event-gateway/subscriptions", kvstore),
		Log:               log,
	}

	ha := &httpapi.HTTPAPI{
		EventTypes:    service,
		Functions:     service,
		Subscriptions: service,
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
