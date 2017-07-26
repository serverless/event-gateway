package tests

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/pubsub"
)

func TestFunctionPubSub(t *testing.T) {
	logCfg := zap.NewDevelopmentConfig()
	logCfg.DisableStacktrace = true
	log, _ := logCfg.Build()

	kv, shutdownGuard := TestingEtcd()

	testAPIServer := newTestAPIServer(kv, log)
	defer testAPIServer.Close()

	router, testRouterServer := newTestRouterServer(kv, log)
	defer testRouterServer.Close()
	router.StartWorkers()

	expected := "😸"

	// register subscriber function
	smileyReceived := make(chan struct{})
	testSubscriberServer := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// read the body, compare the value, close notification chan
			reqBuf, err := ioutil.ReadAll(r.Body)
			if err != nil {
				panic(err)
			}

			if string(reqBuf) == expected {
				close(smileyReceived)
			} else {
				log.Error("received non-smiley!", zap.String("value", fmt.Sprintf("%+v", reqBuf)))
			}
		}))
	defer testSubscriberServer.Close()

	subscriberFnID := functions.FunctionID("smiley subscriber")
	post(testAPIServer.URL+"/v1/functions",
		functions.Function{
			ID: subscriberFnID,
			Provider: &functions.Provider{
				Type: functions.HTTPEndpoint,
				URL:  testSubscriberServer.URL,
			},
		})

	// set up pub/sub
	eventName := "smileys"

	post(testAPIServer.URL+"/v1/subscriptions", pubsub.Subscription{
		FunctionID: subscriberFnID,
		Event:      pubsub.TopicID(eventName),
	})

	wait5Seconds(router.WaitForSubscriber(pubsub.TopicID(eventName)),
		"timed out waiting for subscriber to be configured!")

	emit(testRouterServer.URL, eventName, []byte(expected))

	wait5Seconds(smileyReceived,
		"timed out waiting to receive pub/sub event in subscriber!")

	router.Drain()
	shutdownGuard.ShutdownAndWait()
}

func emit(url, topic string, body []byte) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}

	req.Header.Add("event", topic)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}

func wait5Seconds(ch <-chan struct{}, errMsg string) {
	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		panic(errMsg)
	}
}
