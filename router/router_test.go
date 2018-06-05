package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/plugin"
	"github.com/serverless/event-gateway/router"
	"github.com/serverless/event-gateway/router/mock"
	"github.com/stretchr/testify/assert"
)

func TestRouterServeHTTP_StatusUnavailableWhenDraining(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	target := mock.NewMockTargeter(ctrl)
	router := testrouter(target)
	router.Drain()

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	assert.Equal(t, `{"errors":[{"message":"Service Unavailable"}]}`+"\n", recorder.Body.String())
}

func TestRouterServeHTTP_AllowCORSPreflight(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	target := mock.NewMockTargeter(ctrl)
	router := testrouter(target)

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
}

func TestRouterServeHTTP_ExtractPathFromHostedDomain(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	target := mock.NewMockTargeter(ctrl)
	target.EXPECT().SyncSubscriber(http.MethodGet, "/custom/test", event.TypeName("http.request")).Return(nil).MaxTimes(1)
	target.EXPECT().AsyncSubscribers(http.MethodGet, "/custom/test", event.TypeName("http.request")).Return([]router.AsyncSubscriber{}).MaxTimes(1)
	target.EXPECT().AsyncSubscribers(http.MethodPost, "/", event.SystemEventReceivedType).Return([]router.AsyncSubscriber{}).MaxTimes(1)
	router := testrouter(target)

	req, _ := http.NewRequest(http.MethodGet, "https://custom.slsgateway.com/test", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusAccepted, recorder.Code)
}

func testrouter(target router.Targeter) *router.Router {
	log := zap.NewNop()
	plugins := plugin.NewManager([]string{}, log)
	router := router.New(10, 10, target, plugins, log)
	router.StartWorkers()
	return router
}
