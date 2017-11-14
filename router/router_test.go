package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/internal/cors"
	"github.com/serverless/event-gateway/internal/pathtree"
	"github.com/serverless/event-gateway/plugin"
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
	assert.Equal(t, "Service Unavailable\n", recorder.Body.String())
}

func TestRouterServeHTTP_HTTPEventFunctionNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	target := mock.NewMockTargeter(ctrl)
	target.EXPECT().HTTPBackingFunction(http.MethodGet, "/notfound").Return(nil, pathtree.Params{}, nil).MaxTimes(1)
	target.EXPECT().SubscribersOfEvent("/", event.SystemEventReceivedType).Return([]functions.FunctionID{}).MaxTimes(1)
	router := testrouter(target)

	req, _ := http.NewRequest(http.MethodGet, "/notfound", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Equal(t, "resource not found\n", recorder.Body.String())
}

func TestRouterServeHTTP_InvokeEventFunctionNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	target := mock.NewMockTargeter(ctrl)
	target.EXPECT().Function(functions.FunctionID("testfunc")).Return(nil).MaxTimes(1)
	target.EXPECT().SubscribersOfEvent("/", event.SystemEventReceivedType).Return([]functions.FunctionID{}).MaxTimes(1)
	router := testrouter(target)

	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("event", "invoke")
	req.Header.Set("function-id", "testfunc")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, "unable to look up registered function\n", recorder.Body.String())
}

func TestRouterServeHTTP_ErrorMalformedCustomEventJSONRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	target := mock.NewMockTargeter(ctrl)
	router := testrouter(target)

	req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader("not json"))
	req.Header.Set("content-type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Equal(t, "malformed JSON body\n", recorder.Body.String())
}

func TestRouterServeHTTP_ErrorOnCustomEventEmittedWithNonPostMethod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	target := mock.NewMockTargeter(ctrl)
	target.EXPECT().SubscribersOfEvent("/", event.SystemEventReceivedType).Return([]functions.FunctionID{}).MaxTimes(1)
	router := testrouter(target)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("event", "user.created")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Equal(t, "custom event can be emitted only with POST method\n", recorder.Body.String())
}

func TestRouterServeHTTP_AllowCORSPreflightForHTTPEventWhenConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	target := mock.NewMockTargeter(ctrl)
	id := functions.FunctionID("testid")
	target.EXPECT().HTTPBackingFunction(http.MethodGet, "/test").Return(&id, pathtree.Params{}, &cors.CORS{
		Origins: []string{"http://example.com"},
		Methods: []string{"GET"},
	}).MaxTimes(1)
	target.EXPECT().SubscribersOfEvent("/", event.SystemEventReceivedType).Return([]functions.FunctionID{}).MaxTimes(1)
	router := testrouter(target)

	req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Origin", "http://example.com")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "GET", recorder.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "http://example.com", recorder.Header().Get("Access-Control-Allow-Origin"))
}

func TestRouterServeHTTP_AllowCORSPreflightForCustomEvents(t *testing.T) {
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

func testrouter(target Targeter) *Router {
	log := zap.NewNop()
	plugins := plugin.NewManager([]string{}, log)
	router := New(10, target, plugins, log)
	router.StartWorkers()
	return router
}
