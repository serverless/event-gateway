package router_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/plugin"
	"github.com/serverless/event-gateway/router"
	"github.com/serverless/event-gateway/router/mock"
	"github.com/serverless/event-gateway/subscription/cors"
	"github.com/stretchr/testify/assert"

	httpprovider "github.com/serverless/event-gateway/providers/http"
)

func TestRouterServeHTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	target := mock.NewMockTargeter(ctrl)

	t.Run("status Unavaliable when draining", func(t *testing.T) {
		router := setupTestRouter(target)
		router.Drain()

		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
		assert.Equal(t, `{"errors":[{"message":"Service Unavailable"}]}`+"\n", recorder.Body.String())
	})

	t.Run("allow CORS preflight when configured", func(t *testing.T) {
		config := &cors.CORS{
			AllowedOrigins: []string{"http://example.com"},
			AllowedMethods: []string{"PUT"},
			AllowedHeaders: []string{"*"},
		}
		target.EXPECT().CORS(http.MethodPut, "/").Return(config)
		routera := setupTestRouter(target)

		req, _ := http.NewRequest(http.MethodOptions, "/", nil)
		req.Header.Set("Access-Control-Request-Method", "PUT")
		req.Header.Set("Access-Control-Request-Headers", "x-api-key")
		req.Header.Set("Origin", "http://example.com")
		recorder := httptest.NewRecorder()
		routera.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "", recorder.Header().Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, "X-Api-Key", recorder.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "PUT", recorder.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "http://example.com", recorder.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("reject if system event", func(t *testing.T) {
		target.EXPECT().CORS(gomock.Any(), gomock.Any()).Return(nil)
		router := setupTestRouter(target)

		req, _ := http.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Event", "gateway.something")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("subscriptions handling", func(t *testing.T) {
		space := "default"
		functionID := function.ID("test")
		fn := &function.Function{
			Space:        space,
			ID:           functionID,
			ProviderType: httpprovider.Type,
			Provider:     &httpprovider.HTTP{URL: testHTTPFunction(http.StatusOK, []byte("{}")).URL},
		}
		subscriber := &router.SyncSubscriber{Space: space, FunctionID: functionID}

		t.Run("call sync subscriber", func(t *testing.T) {
			eventType := &event.Type{Space: space, Name: "http.request"}
			target.EXPECT().CORS(gomock.Any(), gomock.Any()).Return(nil)
			target.EXPECT().SyncSubscriber(http.MethodPost, "/", event.TypeHTTPRequest).Return(subscriber).MaxTimes(1)
			target.EXPECT().AsyncSubscribers(gomock.Any(), gomock.Any(), gomock.Any()).Return([]router.AsyncSubscriber{}).AnyTimes()
			target.EXPECT().EventType(space, event.TypeHTTPRequest).Return(eventType)
			target.EXPECT().Function(space, functionID).Return(fn)
			router := setupTestRouter(target)

			req, _ := http.NewRequest(http.MethodPost, "/", nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, http.StatusOK, recorder.Code)
		})

		t.Run("status code and headers based on HTTP response object", func(t *testing.T) {
			httpResponseObject := []byte(`{"statusCode": 206, "headers": {"x-custom": "custom value"}}`)
			fn = &function.Function{
				Space:        space,
				ID:           functionID,
				ProviderType: httpprovider.Type,
				Provider:     &httpprovider.HTTP{URL: testHTTPFunction(http.StatusOK, httpResponseObject).URL},
			}
			eventType := &event.Type{Space: space, Name: "http.request"}
			target.EXPECT().CORS(gomock.Any(), gomock.Any()).Return(nil)
			target.EXPECT().SyncSubscriber(http.MethodPost, "/", event.TypeHTTPRequest).Return(subscriber).MaxTimes(1)
			target.EXPECT().AsyncSubscribers(gomock.Any(), gomock.Any(), gomock.Any()).Return([]router.AsyncSubscriber{}).AnyTimes()
			target.EXPECT().EventType(space, event.TypeHTTPRequest).Return(eventType)
			target.EXPECT().Function(space, functionID).Return(fn)
			router := setupTestRouter(target)

			req, _ := http.NewRequest(http.MethodPost, "/", nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, 206, recorder.Code)
			assert.Equal(t, "custom value", recorder.HeaderMap.Get("x-custom"))
		})

		t.Run("status Internal Server Error if HTTP response object malformed", func(t *testing.T) {
			fn = &function.Function{
				Space:        space,
				ID:           functionID,
				ProviderType: httpprovider.Type,
				Provider:     &httpprovider.HTTP{URL: testHTTPFunction(http.StatusOK, []byte("not JSON")).URL},
			}
			eventType := &event.Type{Space: space, Name: "http.request"}
			target.EXPECT().CORS(gomock.Any(), gomock.Any()).Return(nil)
			target.EXPECT().SyncSubscriber(http.MethodPost, "/", event.TypeHTTPRequest).Return(subscriber).MaxTimes(1)
			target.EXPECT().AsyncSubscribers(gomock.Any(), gomock.Any(), gomock.Any()).Return([]router.AsyncSubscriber{}).AnyTimes()
			target.EXPECT().EventType(space, event.TypeHTTPRequest).Return(eventType)
			target.EXPECT().Function(space, functionID).Return(fn)
			router := setupTestRouter(target)

			req, _ := http.NewRequest(http.MethodPost, "/", nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		})

		t.Run("with authorizer", func(t *testing.T) {
			authorizerID := function.ID("auth")
			authorizer := &function.Function{
				Space:        space,
				ID:           authorizerID,
				ProviderType: httpprovider.Type,
				// function returning nil is not valid authorizer
				Provider: &httpprovider.HTTP{URL: testHTTPFunction(http.StatusOK, nil).URL},
			}
			eventType := &event.Type{Space: space, Name: "http.request", AuthorizerID: &authorizerID}

			t.Run("status Forbidden if authorizer returned nil", func(t *testing.T) {
				target.EXPECT().CORS(gomock.Any(), gomock.Any()).Return(nil)
				target.EXPECT().SyncSubscriber(gomock.Any(), gomock.Any(), gomock.Any()).Return(subscriber).MaxTimes(1)
				target.EXPECT().AsyncSubscribers(gomock.Any(), gomock.Any(), gomock.Any()).Return([]router.AsyncSubscriber{}).AnyTimes()
				target.EXPECT().EventType(gomock.Any(), gomock.Any()).Return(eventType)
				target.EXPECT().Function(space, authorizerID).Return(authorizer)
				router := setupTestRouter(target)

				req, _ := http.NewRequest(http.MethodPost, "/", nil)
				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, req)

				assert.Equal(t, http.StatusForbidden, recorder.Code)
			})

			t.Run("status Forbidden if authorizer returned error", func(t *testing.T) {
				authorizer.Provider = &httpprovider.HTTP{
					URL: testHTTPFunction(http.StatusOK, []byte(`{"error":{"message": "failed"}}`)).URL}
				target.EXPECT().CORS(gomock.Any(), gomock.Any()).Return(nil)
				target.EXPECT().SyncSubscriber(gomock.Any(), gomock.Any(), gomock.Any()).Return(subscriber).MaxTimes(1)
				target.EXPECT().AsyncSubscribers(gomock.Any(), gomock.Any(), gomock.Any()).Return([]router.AsyncSubscriber{}).AnyTimes()
				target.EXPECT().EventType(gomock.Any(), gomock.Any()).Return(eventType)
				target.EXPECT().Function(space, authorizerID).Return(authorizer)
				router := setupTestRouter(target)

				req, _ := http.NewRequest(http.MethodPost, "/", nil)
				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, req)

				assert.Equal(t, http.StatusForbidden, recorder.Code)
			})

			t.Run("include authorizer result in event extensions for event created by EG", func(t *testing.T) {
				targetFunction := httptest.NewServer(http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						cloudEvent := &event.Event{}
						dec := json.NewDecoder(r.Body)
						dec.Decode(cloudEvent)

						assert.Equal(t, "testid", cloudEvent.Extensions["eventgateway"].(map[string]interface{})["authorization"].(map[string]interface{})["principalId"])
					}))

				fn.Provider = &httpprovider.HTTP{
					URL: targetFunction.URL}
				authorizer.Provider = &httpprovider.HTTP{
					URL: testHTTPFunction(http.StatusOK, []byte(`{"authorization":{"principalId": "testid"}}`)).URL}
				target.EXPECT().CORS(gomock.Any(), gomock.Any()).Return(nil)
				target.EXPECT().SyncSubscriber(gomock.Any(), gomock.Any(), gomock.Any()).Return(subscriber).MaxTimes(1)
				target.EXPECT().AsyncSubscribers(gomock.Any(), gomock.Any(), gomock.Any()).Return([]router.AsyncSubscriber{}).AnyTimes()
				target.EXPECT().EventType(gomock.Any(), gomock.Any()).Return(eventType)
				target.EXPECT().Function(space, authorizerID).Return(authorizer)
				target.EXPECT().Function(space, functionID).Return(fn)
				router := setupTestRouter(target)

				req, _ := http.NewRequest(http.MethodPost, "/", nil)
				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, req)
			})

			t.Run("include authorizer result in custom event extensions", func(t *testing.T) {
				targetFunction := httptest.NewServer(http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						cloudEvent := &event.Event{}
						dec := json.NewDecoder(r.Body)
						dec.Decode(cloudEvent)

						assert.Equal(t, "testid", cloudEvent.Extensions["eventgateway"].(map[string]interface{})["authorization"].(map[string]interface{})["principalId"])
					}))

				fn.Provider = &httpprovider.HTTP{
					URL: targetFunction.URL}
				authorizer.Provider = &httpprovider.HTTP{
					URL: testHTTPFunction(http.StatusOK, []byte(`{"authorization":{"principalId": "testid"}}`)).URL}
				eventType := &event.Type{Space: space, Name: "test.event", AuthorizerID: &authorizerID}
				target.EXPECT().CORS(gomock.Any(), gomock.Any()).Return(nil)
				target.EXPECT().SyncSubscriber(gomock.Any(), gomock.Any(), gomock.Any()).Return(subscriber).MaxTimes(1)
				target.EXPECT().AsyncSubscribers(gomock.Any(), gomock.Any(), gomock.Any()).Return([]router.AsyncSubscriber{}).AnyTimes()
				target.EXPECT().EventType(gomock.Any(), gomock.Any()).Return(eventType)
				target.EXPECT().Function(space, authorizerID).Return(authorizer)
				target.EXPECT().Function(space, functionID).Return(fn)
				router := setupTestRouter(target)

				req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewReader(
					[]byte(`{"eventId":"test","eventType":"test.event","eventTypeVersion":"0.1","cloudEventsVersion":"0.1","source":"/"}`)))
				req.Header.Set("content-type", "application/cloudevents+json")
				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, req)
			})
		})
	})
}

func setupTestRouter(target router.Targeter) *router.Router {
	log := zap.NewNop()
	plugins := plugin.NewManager([]string{}, log)
	router := router.New(10, 10, target, plugins, log)
	router.StartWorkers()
	return router
}

func testHTTPFunction(status int, response []byte) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
			w.Write(response)
		}))
}
