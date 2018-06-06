package router_test

import (
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

	t.Run("allow CORS preflight", func(t *testing.T) {
		router := setupTestRouter(target)

		req, _ := http.NewRequest(http.MethodOptions, "/", nil)
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "event")
		req.Header.Set("Origin", "http://example.com")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "true", recorder.Header().Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, "Event", recorder.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "POST", recorder.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "http://example.com", recorder.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("extract path from hosted domain", func(t *testing.T) {
		target.EXPECT().SyncSubscriber(http.MethodGet, "/custom/test", event.TypeName("http.request")).Return(nil).MaxTimes(1)
		target.EXPECT().AsyncSubscribers(http.MethodGet, "/custom/test", event.TypeName("http.request")).Return([]router.AsyncSubscriber{}).MaxTimes(1)
		target.EXPECT().AsyncSubscribers(http.MethodPost, "/", event.SystemEventReceivedType).Return([]router.AsyncSubscriber{}).MaxTimes(1)
		router := setupTestRouter(target)

		req, _ := http.NewRequest(http.MethodGet, "https://custom.slsgateway.com/test", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusAccepted, recorder.Code)
	})

	t.Run("reject if system event", func(t *testing.T) {
		router := setupTestRouter(target)

		req, _ := http.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Event", "gateway.something")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

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

	t.Run("status Forbidden if authorizer call failed", func(t *testing.T) {
		authorizerID := function.ID("auth")
		authorizer := &function.Function{
			Space:        space,
			ID:           authorizerID,
			ProviderType: httpprovider.Type,
			// function returning nil is not valid authorizer
			Provider: &httpprovider.HTTP{URL: testHTTPFunction(http.StatusOK, nil).URL},
		}
		eventType := &event.Type{Space: space, Name: "http.request", AuthorizerID: &authorizerID}
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
