// +build hosted

package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/plugin"
	"github.com/serverless/event-gateway/router"
	"github.com/serverless/event-gateway/router/mock"
	"go.uber.org/zap"
)

func TestHostedRouterServeHTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	target := mock.NewMockTargeter(ctrl)

	t.Run("emit system event 'event.received' on path prefixed with space", func(t *testing.T) {
		target.EXPECT().CORS(gomock.Any(), gomock.Any()).Return(nil)
		target.EXPECT().SyncSubscriber(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		target.EXPECT().AsyncSubscribers(gomock.Any(), gomock.Any(), event.TypeName("http.request")).Return([]router.AsyncSubscriber{})

		target.EXPECT().AsyncSubscribers(http.MethodPost, "/custom/", event.SystemEventReceivedType).Return([]router.AsyncSubscriber{})

		router := setupTestRouter(target)
		req, _ := http.NewRequest(http.MethodGet, "https://custom.slsgateway.com/foo/bar", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
	})

	t.Run("extract path from hosted domain", func(t *testing.T) {
		target.EXPECT().CORS(gomock.Any(), gomock.Any()).Return(nil)
		target.EXPECT().AsyncSubscribers(gomock.Any(), gomock.Any(), event.SystemEventReceivedType).Return([]router.AsyncSubscriber{})

		target.EXPECT().SyncSubscriber(http.MethodGet, "/custom/test", event.TypeName("http.request")).Return(nil)
		target.EXPECT().AsyncSubscribers(http.MethodGet, "/custom/test", event.TypeName("http.request")).Return([]router.AsyncSubscriber{})

		router := setupTestRouter(target)
		req, _ := http.NewRequest(http.MethodGet, "https://custom.slsgateway.com/test", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
	})

	t.Run("if not hosted EG should fallback to full path", func(t *testing.T) {
		target.EXPECT().CORS(gomock.Any(), gomock.Any()).Return(nil)
		target.EXPECT().AsyncSubscribers(gomock.Any(), gomock.Any(), event.SystemEventReceivedType).Return([]router.AsyncSubscriber{})

		target.EXPECT().SyncSubscriber(http.MethodGet, "/foo/bar", event.TypeName("http.request")).Return(nil)
		target.EXPECT().AsyncSubscribers(http.MethodGet, "/foo/bar", event.TypeName("http.request")).Return([]router.AsyncSubscriber{})

		router := setupTestRouter(target)
		req, _ := http.NewRequest(http.MethodGet, "https://127.0.0.1/foo/bar", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
	})

	t.Run("if not hosted EG should fallback to / for 'event.received' system event", func(t *testing.T) {
		target.EXPECT().CORS(gomock.Any(), gomock.Any()).Return(nil)
		target.EXPECT().SyncSubscriber(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		target.EXPECT().AsyncSubscribers(gomock.Any(), gomock.Any(), event.TypeName("http.request")).Return([]router.AsyncSubscriber{})

		target.EXPECT().AsyncSubscribers(http.MethodPost, "/", event.SystemEventReceivedType).Return([]router.AsyncSubscriber{})

		router := setupTestRouter(target)
		req, _ := http.NewRequest(http.MethodGet, "https://127.0.0.1/test", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
	})
}

func setupTestRouter(target router.Targeter) *router.Router {
	log := zap.NewNop()
	plugins := plugin.NewManager([]string{}, log)
	router := router.New(10, 10, target, plugins, log)
	router.StartWorkers()
	return router
}
