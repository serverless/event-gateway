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

	t.Run("call sync subscriber", func(t *testing.T) {
		space := "default"
		funcID := function.ID("test")
		subscriber := &router.SyncSubscriber{
			Space:      space,
			FunctionID: funcID,
		}
		eventType := &event.Type{
			Space: space,
			Name:  "http.request",
		}
		functionHTTP := testHTTPFunction(http.StatusOK, []byte("{}"))
		function := &function.Function{
			Space:        space,
			ID:           funcID,
			ProviderType: httpprovider.Type,
			Provider:     &httpprovider.HTTP{URL: functionHTTP.URL},
		}

		target.EXPECT().SyncSubscriber(http.MethodPost, "/", event.TypeHTTPRequest).Return(subscriber).MaxTimes(1)
		target.EXPECT().AsyncSubscribers(gomock.Any(), gomock.Any(), gomock.Any()).Return([]router.AsyncSubscriber{}).AnyTimes()
		target.EXPECT().EventType(space, event.TypeHTTPRequest).Return(eventType)
		target.EXPECT().Function(space, funcID).Return(function)
		router := setupTestRouter(target)

		req, _ := http.NewRequest(http.MethodPost, "/", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
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
