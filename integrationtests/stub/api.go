package stub

import (
	"net/http/httptest"

	"github.com/docker/libkv/store"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"

	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/internal/kv"
	"github.com/serverless/event-gateway/pubsub"
)

// ConfigAPIServer creates test Configuration API server.
func ConfigAPIServer(kvstore store.Store, log *zap.Logger) *httptest.Server {
	apiRouter := httprouter.New()

	fnsDB := kv.NewPrefixedStore("/serverless-event-gateway/functions", kvstore)
	fns := &functions.Functions{
		DB:  fnsDB,
		Log: log,
	}
	fnsapi := &functions.HTTPAPI{Functions: fns}
	fnsapi.RegisterRoutes(apiRouter)

	ps := &pubsub.PubSub{
		TopicsDB:        kv.NewPrefixedStore("/serverless-event-gateway/topics", kvstore),
		SubscriptionsDB: kv.NewPrefixedStore("/serverless-event-gateway/subscriptions", kvstore),
		EndpointsDB:     kv.NewPrefixedStore("/serverless-event-gateway/endpoints", kvstore),
		FunctionsDB:     fnsDB,
		Log:             log,
	}
	psapi := &pubsub.HTTPAPI{PubSub: ps}
	psapi.RegisterRoutes(apiRouter)

	return httptest.NewServer(apiRouter)
}
