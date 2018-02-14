package httpapi_test

import (
	"encoding/json"
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

func TestRegisterFunction_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	router, functions, _ := setup(ctrl)

	returned := &function.Function{ID: function.ID("func1"), Space: "default"}
	functions.EXPECT().GetFunction("default", function.ID("func1")).Return(returned, nil)

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/functions/default/func1", nil)
	router.ServeHTTP(resp, req)

	f := &function.Function{}
	json.Unmarshal(resp.Body.Bytes(), &f)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "default", f.Space)
	assert.Equal(t, function.ID("func1"), f.ID)
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
