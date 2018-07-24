// +build hosted

package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/router"
	"github.com/serverless/event-gateway/router/mock"
	"github.com/stretchr/testify/assert"
)

func TestHostedRouterServeHTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	target := mock.NewMockTargeter(ctrl)

	t.Run("extract path from hosted domain", func(t *testing.T) {
		target.EXPECT().CORS(gomock.Any(), gomock.Any()).Return(nil)
		target.EXPECT().SyncSubscriber(http.MethodGet, "/custom/test", event.TypeName("http.request")).Return(nil).MaxTimes(1)
		target.EXPECT().AsyncSubscribers(http.MethodGet, "/custom/test", event.TypeName("http.request")).Return([]router.AsyncSubscriber{}).MaxTimes(1)
		target.EXPECT().AsyncSubscribers(http.MethodPost, "/", event.SystemEventReceivedType).Return([]router.AsyncSubscriber{}).MaxTimes(1)
		router := setupTestRouter(target)

		req, _ := http.NewRequest(http.MethodGet, "https://custom.slsgateway.com/test", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusAccepted, recorder.Code)
	})
}
