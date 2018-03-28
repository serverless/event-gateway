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
	"github.com/serverless/event-gateway/subscription"
	"github.com/serverless/event-gateway/httpapi"
	"github.com/serverless/event-gateway/mock"
	"github.com/stretchr/testify/assert"

	httpprovider "github.com/serverless/event-gateway/providers/http"
)

func TestGetFunction_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, functions, _ := setup(ctrl)

	returned := &function.Function{
		Space:        "default",
		ID:           function.ID("func1"),
		ProviderType: httpprovider.Type,
		Provider:     &httpprovider.HTTP{URL: "http://example.com"},
	}
	functions.EXPECT().GetFunction("default", function.ID("func1")).Return(returned, nil)

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/spaces/default/functions/func1", nil)
	router.ServeHTTP(resp, req)

	fn := &function.Function{}
	json.Unmarshal(resp.Body.Bytes(), fn)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "default", fn.Space)
	assert.Equal(t, function.ID("func1"), fn.ID)
	assert.Equal(t, httpprovider.Type, fn.ProviderType)
	assert.Equal(t, &httpprovider.HTTP{URL: "http://example.com"}, fn.Provider)
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

	returned := function.Functions{{
		ID:           function.ID("func1"),
		Space:        "default",
		ProviderType: httpprovider.Type,
		Provider:     &httpprovider.HTTP{},
	}}
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
		ID:           function.ID("func1"),
		Space:        "test1",
		ProviderType: httpprovider.Type,
		Provider: &httpprovider.HTTP{
			URL: "http://example.com",
		},
	}
	functions.EXPECT().RegisterFunction(fn).Return(fn, nil)

	resp := httptest.NewRecorder()
	payload := bytes.NewReader([]byte(`
		{"functionId":"func1","space":"test1","type":"http","provider":{"url":"http://example.com"}}
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

func TestRegisterFunction_AlreadyRegistered(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, functions, _ := setup(ctrl)

	fn := &function.Function{
		ID:           function.ID("func1"),
		Space:        "default",
		ProviderType: httpprovider.Type,
		Provider:     &httpprovider.HTTP{URL: "http://test.com"},
	}
	functions.EXPECT().RegisterFunction(fn).Return(nil, &function.ErrFunctionAlreadyRegistered{ID: function.ID("func1")})

	resp := httptest.NewRecorder()
	payload := bytes.NewReader([]byte(`{"functionID":"func1","type":"http","provider":{"url":"http://test.com"}}}`))
	req, _ := http.NewRequest(http.MethodPost, "/v1/spaces/default/functions", payload)
	router.ServeHTTP(resp, req)

	httpresp := &httpapi.Response{}
	json.Unmarshal(resp.Body.Bytes(), httpresp)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Equal(t, `Function "func1" already registered.`, httpresp.Errors[0].Message)
}

func TestRegisterFunction_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, functions, _ := setup(ctrl)

	fn := &function.Function{
		ID:           function.ID("/"),
		Space:        "default",
		ProviderType: httpprovider.Type,
		Provider:     &httpprovider.HTTP{URL: "http://test.com"},
	}
	functions.EXPECT().RegisterFunction(fn).Return(nil, &function.ErrFunctionValidation{Message: "wrong function ID format"})

	resp := httptest.NewRecorder()
	payload := bytes.NewReader([]byte(`{"functionID":"/","type":"http","provider":{"url":"http://test.com"}}}`))
	req, _ := http.NewRequest(http.MethodPost, "/v1/spaces/default/functions", payload)
	router.ServeHTTP(resp, req)

	httpresp := &httpapi.Response{}
	json.Unmarshal(resp.Body.Bytes(), httpresp)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Equal(t, "Function doesn't validate. Validation error: wrong function ID format", httpresp.Errors[0].Message)
}

func TestRegisterFunction_MalformedJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, _, _ := setup(ctrl)

	resp := httptest.NewRecorder()
	payload := bytes.NewReader([]byte(`{`))
	req, _ := http.NewRequest(http.MethodPost, "/v1/spaces/default/functions", payload)
	router.ServeHTTP(resp, req)

	httpresp := &httpapi.Response{}
	json.Unmarshal(resp.Body.Bytes(), httpresp)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Equal(t, "Function doesn't validate. Validation error: unexpected EOF", httpresp.Errors[0].Message)
}

func TestRegisterFunction_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, functions, _ := setup(ctrl)

	fn := &function.Function{
		ID:           function.ID("func1"),
		Space:        "default",
		ProviderType: httpprovider.Type,
		Provider:     &httpprovider.HTTP{URL: "http://example.com"},
	}
	functions.EXPECT().RegisterFunction(fn).Return(nil, errors.New("processing error"))

	resp := httptest.NewRecorder()
	payload := bytes.NewReader([]byte(`{"functionID":"func1","type":"http","provider":{"url":"http://example.com"}}}`))
	req, _ := http.NewRequest(http.MethodPost, "/v1/spaces/default/functions", payload)
	router.ServeHTTP(resp, req)

	httpresp := &httpapi.Response{}
	json.Unmarshal(resp.Body.Bytes(), httpresp)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, `processing error`, httpresp.Errors[0].Message)
}

func TestDeleteFunction_BadRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, functions, _ := setup(ctrl)

	functions.EXPECT().DeleteFunction("default", function.ID("func1")).Return(&function.ErrFunctionHasSubscriptionsError{})

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/v1/spaces/default/functions/func1", nil)
	router.ServeHTTP(resp, req)

	httpresp := &httpapi.Response{}
	json.Unmarshal(resp.Body.Bytes(), httpresp)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Equal(t, "Function cannot be deleted because it's subscribed to a least one event.", httpresp.Errors[0].Message)
}

func TestDeleteFunction_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, functions, _ := setup(ctrl)

	functions.EXPECT().DeleteFunction("default", function.ID("func1")).Return(&function.ErrFunctionNotFound{ID: function.ID("testid")})

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/v1/spaces/default/functions/func1", nil)
	router.ServeHTTP(resp, req)

	httpresp := &httpapi.Response{}
	json.Unmarshal(resp.Body.Bytes(), httpresp)
	assert.Equal(t, http.StatusNotFound, resp.Code)
	assert.Equal(t, `Function "testid" not found.`, httpresp.Errors[0].Message)
}

func TestDeleteFunction_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, functions, _ := setup(ctrl)

	functions.EXPECT().DeleteFunction("default", function.ID("func1")).Return(errors.New("internal error"))

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/v1/spaces/default/functions/func1", nil)
	router.ServeHTTP(resp, req)

	httpresp := &httpapi.Response{}
	json.Unmarshal(resp.Body.Bytes(), httpresp)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, "internal error", httpresp.Errors[0].Message)
}

func TestDeleteFunction_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, functions, _ := setup(ctrl)

	functions.EXPECT().DeleteFunction("default", function.ID("func1")).Return(nil)

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/v1/spaces/default/functions/func1", nil)
	router.ServeHTTP(resp, req)

	httpresp := &httpapi.Response{}
	json.Unmarshal(resp.Body.Bytes(), httpresp)
	assert.Equal(t, http.StatusNoContent, resp.Code)
}

func TestUpdateSubscription_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, _, subscriptions := setup(ctrl)

	returned := &subscription.Subscription{
		Space:        "default",
		ID:           subscription.ID("http,GET,%2F"),
		Event:        "http",
		FunctionID:   "func",
		Method:       "GET",
		Path:         "/",
		CORS:         &subscription.CORS{
		    Origins:          []string{"*"},
		    Methods:          []string{"HEAD", "GET", "POST"},
		    Headers:          []string{"Origin", "Accept", "Content-Type"},
		    AllowCredentials: false,
		},
	}
	subscriptions.EXPECT().UpdateSubscription(subscription.ID("http,GET,%2F"), returned).Return(returned, nil)

	resp := httptest.NewRecorder()
	payload := bytes.NewReader([]byte(`
	    {"space":"default","subscriptionId":"http,GET,%2F","event":"http","functionId":"func","method":"GET","path":"/","cors":{"origins":["*"],"methods":["HEAD","GET","POST"],"headers":["Origin","Accept","Content-Type"],"allowCredentials":false}}
		`))
	req, _ := http.NewRequest(http.MethodPut, "/v1/spaces/default/subscriptions/http,GET,%2F", payload)
	router.ServeHTTP(resp, req)

	sub := &subscription.Subscription{}
	json.Unmarshal(resp.Body.Bytes(), sub)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "default", sub.Space)
	assert.Equal(t, subscription.ID("http,GET,%2F"), sub.ID)
}

func TestUpdateSubscription_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, _, _:= setup(ctrl)

	resp := httptest.NewRecorder()
	payload := bytes.NewReader([]byte(`{"name":"te`))
	req, _ := http.NewRequest(http.MethodPut, "/v1/spaces/default/subscriptions/http,GET,%2F", payload)
	router.ServeHTTP(resp, req)

	sub := &subscription.Subscription{}
	json.Unmarshal(resp.Body.Bytes(), sub)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestUpdateSubscription_InvalidSubscriptionUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, _, subscriptions := setup(ctrl)

	input := &subscription.Subscription{
		Space:        "default",
		ID:           subscription.ID("http,GET,%2F"),
		Event:        "http",
		FunctionID:   "func2",
		Method:       "GET",
		Path:         "/",
		CORS:         &subscription.CORS{
		    Origins:          []string{"*"},
		    Methods:          []string{"HEAD", "GET", "POST"},
		    Headers:          []string{"Origin", "Accept", "Content-Type"},
		    AllowCredentials: false,
		},
	}
	subscriptions.EXPECT().UpdateSubscription(subscription.ID("http,GET,%2F"), input).Return(nil, &subscription.ErrInvalidSubscriptionUpdate{Field: "FunctionID"})

	resp := httptest.NewRecorder()
	payload := bytes.NewReader([]byte(`
	    {"space":"default","subscriptionId":"http,GET,%2F","event":"http","functionId":"func2","method":"GET","path":"/","cors":{"origins":["*"],"methods":["HEAD","GET","POST"],"headers":["Origin","Accept","Content-Type"],"allowCredentials":false}}
		`))
	req, _ := http.NewRequest(http.MethodPut, "/v1/spaces/default/subscriptions/http,GET,%2F", payload)
	router.ServeHTTP(resp, req)

	httpresp := &httpapi.Response{}
	json.Unmarshal(resp.Body.Bytes(), httpresp)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Equal(t, `Invalid update. 'FunctionID' of existing subscription cannot be updated.`, httpresp.Errors[0].Message)
}

func TestUpdateSubscription_SubscriptionNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, _, subscriptions := setup(ctrl)

	input := &subscription.Subscription{
		Space:        "default",
		ID:           subscription.ID("http,GET,%2F"),
		Event:        "http",
		FunctionID:   "func",
		Method:       "GET",
		Path:         "/",
		CORS:         &subscription.CORS{
		    Origins:          []string{"*"},
		    Methods:          []string{"HEAD", "GET", "POST"},
		    Headers:          []string{"Origin", "Accept", "Content-Type"},
		    AllowCredentials: false,
		},
	}
	subscriptions.EXPECT().UpdateSubscription(subscription.ID("http,GET,%2F"), input).Return(nil, &subscription.ErrSubscriptionNotFound{ID: subscription.ID("http,GET,%2F")})

	resp := httptest.NewRecorder()
	payload := bytes.NewReader([]byte(`
	    {"space":"default","subscriptionId":"http,GET,%2F","event":"http","functionId":"func","method":"GET","path":"/","cors":{"origins":["*"],"methods":["HEAD","GET","POST"],"headers":["Origin","Accept","Content-Type"],"allowCredentials":false}}
		`))
	req, _ := http.NewRequest(http.MethodPut, "/v1/spaces/default/subscriptions/http,GET,%2F", payload)
	router.ServeHTTP(resp, req)

	httpresp := &httpapi.Response{}
	json.Unmarshal(resp.Body.Bytes(), httpresp)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Equal(t, `Subscription "http,GET,%2F" not found.`, httpresp.Errors[0].Message)
}

func TestUpdateSubscription_FunctionNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, _, subscriptions := setup(ctrl)

	input := &subscription.Subscription{
		Space:        "default",
		ID:           subscription.ID("http,GET,%2F"),
		Event:        "http",
		FunctionID:   "func",
		Method:       "GET",
		Path:         "/",
		CORS:         &subscription.CORS{
		    Origins:          []string{"*"},
		    Methods:          []string{"HEAD", "GET", "POST"},
		    Headers:          []string{"Origin", "Accept", "Content-Type"},
		    AllowCredentials: false,
		},
	}
	subscriptions.EXPECT().UpdateSubscription(subscription.ID("http,GET,%2F"), input).Return(nil, &function.ErrFunctionNotFound{ID: function.ID("func")})

	resp := httptest.NewRecorder()
	payload := bytes.NewReader([]byte(`
	    {"space":"default","subscriptionId":"http,GET,%2F","event":"http","functionId":"func","method":"GET","path":"/","cors":{"origins":["*"],"methods":["HEAD","GET","POST"],"headers":["Origin","Accept","Content-Type"],"allowCredentials":false}}
		`))
	req, _ := http.NewRequest(http.MethodPut, "/v1/spaces/default/subscriptions/http,GET,%2F", payload)
	router.ServeHTTP(resp, req)

	httpresp := &httpapi.Response{}
	json.Unmarshal(resp.Body.Bytes(), httpresp)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Equal(t, `Function "func" not found.`, httpresp.Errors[0].Message)
}

func TestUpdateSubscription_SubscriptionValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, _, subscriptions := setup(ctrl)

	input := &subscription.Subscription{
		Space:        "default",
		ID:           subscription.ID("http,GET,%2F"),
		FunctionID:   "func",
		Method:       "GET",
		Path:         "/",
		CORS:         &subscription.CORS{
		    Origins:          []string{"*"},
		    Methods:          []string{"HEAD", "GET", "POST"},
		    Headers:          []string{"Origin", "Accept", "Content-Type"},
		    AllowCredentials: false,
		},
	}
	subscriptions.EXPECT().UpdateSubscription(subscription.ID("http,GET,%2F"), input).Return(nil, &subscription.ErrSubscriptionValidation{Message: "" })

	resp := httptest.NewRecorder()
	payload := bytes.NewReader([]byte(`
	    {"space":"default","subscriptionId":"http,GET,%2F","functionId":"func","method":"GET","path":"/","cors":{"origins":["*"],"methods":["HEAD","GET","POST"],"headers":["Origin","Accept","Content-Type"],"allowCredentials":false}}
		`))
	req, _ := http.NewRequest(http.MethodPut, "/v1/spaces/default/subscriptions/http,GET,%2F", payload)
	router.ServeHTTP(resp, req)

	httpresp := &httpapi.Response{}
	json.Unmarshal(resp.Body.Bytes(), httpresp)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, httpresp.Errors[0].Message, "Subscription doesn't validate. Validation error")
}

func TestUpdateSubscription_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, _, subscriptions := setup(ctrl)

	input := &subscription.Subscription{
		Space:        "default",
		ID:           subscription.ID("http,GET,%2F"),
		Event:        "http",
		FunctionID:   "func",
		Method:       "GET",
		Path:         "/",
		CORS:         &subscription.CORS{
		    Origins:          []string{"*"},
		    Methods:          []string{"HEAD", "GET", "POST"},
		    Headers:          []string{"Origin", "Accept", "Content-Type"},
		    AllowCredentials: false,
		},
	}
	subscriptions.EXPECT().UpdateSubscription(subscription.ID("http,GET,%2F"), input).Return(nil, errors.New("processing failed"))

	resp := httptest.NewRecorder()
	payload := bytes.NewReader([]byte(`
	    {"space":"default","subscriptionId":"http,GET,%2F","event":"http","functionId":"func","method":"GET","path":"/","cors":{"origins":["*"],"methods":["HEAD","GET","POST"],"headers":["Origin","Accept","Content-Type"],"allowCredentials":false}}
		`))
	req, _ := http.NewRequest(http.MethodPut, "/v1/spaces/default/subscriptions/http,GET,%2F", payload)
	router.ServeHTTP(resp, req)

	httpresp := &httpapi.Response{}
	json.Unmarshal(resp.Body.Bytes(), httpresp)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, "processing failed", httpresp.Errors[0].Message)
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
