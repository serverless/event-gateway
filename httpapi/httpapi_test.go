package httpapi_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/julienschmidt/httprouter"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/httpapi"
	"github.com/serverless/event-gateway/mock"
	"github.com/stretchr/testify/assert"
)

func TestGetFunction_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, functions, _ := setup(ctrl)

	returned := &function.Function{ID: function.ID("func1"), Space: "default"}
	functions.EXPECT().GetFunction("default", function.ID("func1")).Return(returned, nil)

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/spaces/default/functions/func1", nil)
	router.ServeHTTP(resp, req)

	fn := &function.Function{}
	json.Unmarshal(resp.Body.Bytes(), fn)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "default", fn.Space)
	assert.Equal(t, function.ID("func1"), fn.ID)
}

func TestGetFunction_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, functions, _ := setup(ctrl)

	returned := &function.ErrFunctionNotFound{ID: function.ID("func1")}
	functions.EXPECT().GetFunction("default", function.ID("func1")).Return(nil, returned)

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/spaces/default/functions/func1", nil)
	router.ServeHTTP(resp, req)

	httpresp := &httpapi.Response{}
	json.Unmarshal(resp.Body.Bytes(), httpresp)
	assert.Equal(t, http.StatusNotFound, resp.Code)
	assert.Equal(t, `Function "func1" not found.`, httpresp.Errors[0].Message)
}

func TestGetFunction_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, functions, _ := setup(ctrl)

	functions.EXPECT().GetFunction("default", function.ID("func1")).Return(nil, errors.New("processing failed"))

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/spaces/default/functions/func1", nil)
	router.ServeHTTP(resp, req)

	httpresp := &httpapi.Response{}
	json.Unmarshal(resp.Body.Bytes(), httpresp)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, "processing failed", httpresp.Errors[0].Message)
}

func TestGetFunctions_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, functions, _ := setup(ctrl)

	returned := function.Functions{{ID: function.ID("func1"), Space: "default"}}
	functions.EXPECT().GetFunctions("default").Return(returned, nil)

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/spaces/default/functions", nil)
	router.ServeHTTP(resp, req)

	fns := &httpapi.FunctionsResponse{}
	json.Unmarshal(resp.Body.Bytes(), fns)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "default", fns.Functions[0].Space)
	assert.Equal(t, function.ID("func1"), fns.Functions[0].ID)
}

func TestGetFunctions_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, functions, _ := setup(ctrl)

	functions.EXPECT().GetFunctions("default").Return(nil, errors.New("processing failed"))

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/spaces/default/functions", nil)
	router.ServeHTTP(resp, req)

	httpresp := &httpapi.Response{}
	json.Unmarshal(resp.Body.Bytes(), httpresp)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, "processing failed", httpresp.Errors[0].Message)
}

func TestRegisterFunction_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, functions, _ := setup(ctrl)

	fn := &function.Function{
		ID:    function.ID("func1"),
		Space: "test1",
		Provider: &function.Provider{
			Type: function.HTTPEndpoint,
			URL:  "http://example.com",
		},
	}
	functions.EXPECT().RegisterFunction(fn).Return(fn, nil)

	resp := httptest.NewRecorder()
	payload := bytes.NewReader([]byte(`
		{"functionID":"func1", "space":"test1", "provider":{"type":"http", "url":"http://example.com"}}
		`))
	req, _ := http.NewRequest(http.MethodPost, "/v1/spaces/test1/functions", payload)
	router.ServeHTTP(resp, req)

	fn = &function.Function{}
	json.Unmarshal(resp.Body.Bytes(), fn)
	assert.Equal(t, http.StatusCreated, resp.Code)
	assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
	assert.Equal(t, function.ID("func1"), fn.ID)
	assert.Equal(t, "test1", fn.Space)
}

func TestRegisterFunction_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, functions, _ := setup(ctrl)

	fn := &function.Function{
		ID:    function.ID("func1"),
		Space: "default",
	}
	functions.EXPECT().RegisterFunction(fn).Return(nil, &function.ErrFunctionValidation{Message: "no provider"})

	resp := httptest.NewRecorder()
	payload := bytes.NewReader([]byte(`{"functionID":"func1"}}`))
	req, _ := http.NewRequest(http.MethodPost, "/v1/spaces/default/functions", payload)
	router.ServeHTTP(resp, req)

	httpresp := &httpapi.Response{}
	json.Unmarshal(resp.Body.Bytes(), httpresp)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Equal(t, `Function doesn't validate. Validation error: "no provider"`, httpresp.Errors[0].Message)
}

func TestRegisterFunction_AlreadyRegistered(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, functions, _ := setup(ctrl)

	fn := &function.Function{
		ID:    function.ID("func1"),
		Space: "default",
	}
	functions.EXPECT().RegisterFunction(fn).Return(nil, &function.ErrFunctionAlreadyRegistered{ID: function.ID("func1")})

	resp := httptest.NewRecorder()
	payload := bytes.NewReader([]byte(`{"functionID":"func1"}}`))
	req, _ := http.NewRequest(http.MethodPost, "/v1/spaces/default/functions", payload)
	router.ServeHTTP(resp, req)

	httpresp := &httpapi.Response{}
	json.Unmarshal(resp.Body.Bytes(), httpresp)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Equal(t, `Function "func1" already registered.`, httpresp.Errors[0].Message)
}

func TestRegisterFunction_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, functions, _ := setup(ctrl)

	fn := &function.Function{
		ID:    function.ID("func1"),
		Space: "default",
	}
	functions.EXPECT().RegisterFunction(fn).Return(nil, errors.New("processing error"))

	resp := httptest.NewRecorder()
	payload := bytes.NewReader([]byte(`{"functionID":"func1"}}`))
	req, _ := http.NewRequest(http.MethodPost, "/v1/spaces/default/functions", payload)
	router.ServeHTTP(resp, req)

	httpresp := &httpapi.Response{}
	json.Unmarshal(resp.Body.Bytes(), httpresp)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, `processing error`, httpresp.Errors[0].Message)
}

func setup(ctrl *gomock.Controller) (
	*httprouter.Router,
	*mock.MockFunctionService,
	*mock.MockSubscriptionService,
) {
	router := httprouter.New()
	functions := mock.NewMockFunctionService(ctrl)
	subscriptions := mock.NewMockSubscriptionService(ctrl)

	httpapi := &httpapi.HTTPAPI{
		Functions:     functions,
		Subscriptions: subscriptions,
	}
	httpapi.RegisterRoutes(router)

	return router, functions, subscriptions
}
