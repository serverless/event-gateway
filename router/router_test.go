package router_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/internal/pathtree"
	"github.com/serverless/event-gateway/plugin"
	httpprovider "github.com/serverless/event-gateway/providers/http"
	"github.com/serverless/event-gateway/router"
	"github.com/serverless/event-gateway/router/mock"
	"github.com/serverless/event-gateway/subscription"
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

func TestRouterServeHTTP_HTTPEventFunctionNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	target := mock.NewMockTargeter(ctrl)
	target.EXPECT().HTTPBackingFunction(http.MethodGet, "/notfound").Return("", nil, pathtree.Params{}, nil).MaxTimes(1)
	target.EXPECT().SubscribersOfEvent("/", event.SystemEventReceivedType).Return([]router.FunctionInfo{}).MaxTimes(1)
	router := testrouter(target)

	req, _ := http.NewRequest(http.MethodGet, "/notfound", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Equal(t, `{"errors":[{"message":"resource not found"}]}`+"\n", recorder.Body.String())
}

func TestRouterServeHTTP_InvokeEventFunctionNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	target := mock.NewMockTargeter(ctrl)
	target.EXPECT().Function("default", function.ID("testfunc")).Return(nil).MaxTimes(1)
	target.EXPECT().InvokableFunction("/", "default", function.ID("testfunc")).Return(true).MaxTimes(1)
	target.EXPECT().SubscribersOfEvent("/", gomock.Any()).Return([]router.FunctionInfo{}).MaxTimes(2)
	router := testrouter(target)

	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("event", "invoke")
	req.Header.Set("space", "default")
	req.Header.Set("function-id", "testfunc")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, `{"errors":[{"message":"Function call failed. Please check logs."}]}`+"\n", recorder.Body.String())
}

func TestRouterServeHTTP_InvokeEventDefaultSpace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	target := mock.NewMockTargeter(ctrl)
	target.EXPECT().Function("default", function.ID("testfunc")).Return(nil).MaxTimes(1)
	target.EXPECT().InvokableFunction("/", "default", function.ID("testfunc")).Return(true).MaxTimes(1)
	target.EXPECT().SubscribersOfEvent("/", gomock.Any()).Return([]router.FunctionInfo{}).MaxTimes(2)
	router := testrouter(target)

	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("event", "invoke")
	req.Header.Set("function-id", "testfunc")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
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
	assert.Equal(t, `{"errors":[{"message":"malformed JSON body"}]}`+"\n", recorder.Body.String())
}

func TestRouterServeHTTP_Encoding(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	tests := []map[string]string{
		{
			"body": "some=thing",
			"expected": "c29tZT10aGluZw==",
			"content-type": "",
		},
		{
			"body": "some=thing",
			"expected": "some=thing",
			"content-type": "application/x-www-form-urlencoded; charset=utf-8",
		},
		{
			"body": "some=thing",
			"expected": "some=thing",
			"content-type": "application/x-www-form-urlencoded",
		},
		{
			"body": "--X-INSOMNIA-BOUNDARY\r\nContent-Disposition: form-data; name=\"some\"\r\n\r\nthing\r\n--X-INSOMNIA-BOUNDARY--\r\n",
			"expected": "--X-INSOMNIA-BOUNDARY\r\nContent-Disposition: form-data; name=\"some\"\r\n\r\nthing\r\n--X-INSOMNIA-BOUNDARY--\r\n",
			"content-type": "multipart/form-data; boundary=X-INSOMNIA-BOUNDARY",
		},
	}
	for _, test := range tests {
		testListServer := httptest.NewServer(http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				testevent := event.Event{
					Data: event.HTTPEvent{},
				}
				json.NewDecoder(r.Body).Decode(&testevent)

				assert.Equal(t, test["expected"], testevent.Data.(map[string]interface{})["body"])
			}))
		defer testListServer.Close()
		target := mock.NewMockTargeter(ctrl)
		someFunc := function.Function{
			Space:        "",
			ID:           "somefunc",
			ProviderType: httpprovider.Type,
			Provider: httpprovider.HTTP{
				URL: testListServer.URL,
			},
		}
		target.EXPECT().HTTPBackingFunction(http.MethodPost, "/").Return("", &someFunc.ID, pathtree.Params{}, nil)
		target.EXPECT().Function("", someFunc.ID).Return(&someFunc)
		target.EXPECT().SubscribersOfEvent(gomock.Any(), gomock.Any()).Return([]router.FunctionInfo{}).MaxTimes(3)
		router := testrouter(target)

		req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(test["body"]))
		req.Header.Set("content-type", test["content-type"])
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
	}
}

func TestRouterServeHTTP_ErrorOnCustomEventEmittedWithNonPostMethod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	target := mock.NewMockTargeter(ctrl)
	target.EXPECT().SubscribersOfEvent("/", event.SystemEventReceivedType).Return([]router.FunctionInfo{}).MaxTimes(1)
	router := testrouter(target)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("event", "user.created")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Equal(t, `{"errors":[{"message":"custom event can be emitted only with POST method"}]}`+"\n", recorder.Body.String())
}

func TestRouterServeHTTP_AllowCORSPreflightForHTTPEventWhenConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	target := mock.NewMockTargeter(ctrl)
	id := function.ID("testid")
	target.EXPECT().HTTPBackingFunction(http.MethodGet, "/test").Return("default", &id, pathtree.Params{}, &subscription.CORS{
		Origins: []string{"http://example.com"},
		Methods: []string{"GET"},
	}).MaxTimes(1)
	target.EXPECT().SubscribersOfEvent("/", event.SystemEventReceivedType).Return([]router.FunctionInfo{}).MaxTimes(1)
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

func TestRouterServeHTTP_ExtractPathFromHostedDomain(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	target := mock.NewMockTargeter(ctrl)
	target.EXPECT().HTTPBackingFunction(http.MethodGet, "/custom/test").Return("", nil, pathtree.Params{}, &subscription.CORS{}).MaxTimes(1)
	target.EXPECT().SubscribersOfEvent("/", event.SystemEventReceivedType).Return([]router.FunctionInfo{}).MaxTimes(1)
	router := testrouter(target)

	req, _ := http.NewRequest(http.MethodGet, "https://custom.slsgateway.com/test", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func testrouter(target router.Targeter) *router.Router {
	log := zap.NewNop()
	plugins := plugin.NewManager([]string{}, log)
	router := router.New(10, 10, target, plugins, log)
	router.StartWorkers()
	return router
}
