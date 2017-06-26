package tests

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/serverless/event-gateway/endpoints"
	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/pubsub"
)

func wait5Seconds(ch <-chan struct{}, errMsg string) {
	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		panic(errMsg)
	}
}

func TestFunctionPubSub(t *testing.T) {
	logCfg := zap.NewDevelopmentConfig()
	logCfg.DisableStacktrace = true
	log, _ := logCfg.Build()

	kv, shutdown, shutdownComplete := TestingEtcd()

	testAPIServer := newTestAPIServer(kv, log)
	defer testAPIServer.Close()

	router, testRouterServer := newTestRouterServer(kv, log)
	defer testRouterServer.Close()
	router.StartWorkers()

	// register endpoint function
	expected := "😸"

	testTargetServer := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, expected)
		}))
	defer testTargetServer.Close()

	publisherFnID := functions.FunctionID("super smiley function")
	post(testAPIServer.URL+"/v0/gateway/api/function",
		functions.Function{
			ID: publisherFnID,
			HTTP: &functions.HTTPProperties{
				URL: testTargetServer.URL,
			},
		})

	post(testAPIServer.URL+"/v0/gateway/api/endpoint", endpoints.Endpoint{
		FunctionID: publisherFnID,
		Method:     "POST",
		Path:       "/smilez",
	})

	wait5Seconds(router.WaitForEndpoint(endpoints.EndpointID("POST-smilez")),
		"timed out waiting for endpoint to be configured!")

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
	post(testAPIServer.URL+"/v0/gateway/api/function",
		functions.Function{
			ID: subscriberFnID,
			HTTP: &functions.HTTPProperties{
				URL: testSubscriberServer.URL,
			},
		})

	// set up pub/sub
	topicName := "smileys"
	post(testAPIServer.URL+"/v0/gateway/api/topic",
		pubsub.Topic{
			ID: pubsub.TopicID(topicName),
		})

	post(testAPIServer.URL+"/v0/gateway/api/topic/"+topicName+"/subscription",
		pubsub.Subscription{
			FunctionID: subscriberFnID,
		})

	wait5Seconds(router.WaitForSubscriber(pubsub.TopicID(topicName)),
		"timed out waiting for subscriber to be configured!")

	post(testAPIServer.URL+"/v0/gateway/api/topic/"+topicName+"/publisher",
		pubsub.Publisher{
			Type:       "output",
			FunctionID: publisherFnID,
		})

	wait5Seconds(router.WaitForFnPublisher(publisherFnID, "output"),
		"timed out waiting for publisher to be configured!")

	// trigger the endpoint function and wait for the
	// subscriber function to receive the callback.
	res := get(testRouterServer.URL + "/smilez")

	if res != expected {
		panic("returned value was not \"" + expected +
			"\", unexpected value: \"" + res + "\"")
	}

	wait5Seconds(smileyReceived,
		"timed out waiting to receive pub/sub event in subscriber!")

	close(shutdown)
	router.Drain()
	<-shutdownComplete
}
