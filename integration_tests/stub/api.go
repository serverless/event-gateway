package stub

import (
	"net/http/httptest"

	"github.com/docker/libkv/store"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"

	"github.com/serverless/event-gateway/db"
	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/pubsub"
)

// ConfigAPIServer creates test Configuration API server.
func ConfigAPIServer(kv store.Store, log *zap.Logger) *httptest.Server {
	apiRouter := httprouter.New()

	fnsDB := db.NewPrefixedStore("/serverless-event-gateway/functions", kv)
	fns := &functions.Functions{
		DB:     fnsDB,
		Logger: log,
	}
	fnsapi := &functions.HTTPAPI{Functions: fns}
	fnsapi.RegisterRoutes(apiRouter)

	ps := &pubsub.PubSub{
		TopicsDB:        db.NewPrefixedStore("/serverless-event-gateway/topics", kv),
		SubscriptionsDB: db.NewPrefixedStore("/serverless-event-gateway/subscriptions", kv),
		EndpointsDB:     db.NewPrefixedStore("/serverless-event-gateway/endpoints", kv),
		FunctionsDB:     fnsDB,
		Logger:          log,
	}
	psapi := &pubsub.HTTPAPI{PubSub: ps}
	psapi.RegisterRoutes(apiRouter)

	return httptest.NewServer(apiRouter)
}
