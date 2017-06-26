package tests

import (
	"net/http/httptest"

	"github.com/docker/libkv/store"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"

	"github.com/serverless/event-gateway/db"
	"github.com/serverless/event-gateway/endpoints"
	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/pubsub"
)

func newTestAPIServer(kv store.Store, log *zap.Logger) *httptest.Server {
	fnsDB := db.NewPrefixedStore("/serverless-gateway/functions", kv)
	fns := &functions.Functions{
		DB:     fnsDB,
		Logger: log,
	}
	fnsapi := &functions.HTTPAPI{Functions: fns}

	apiRouter := httprouter.New()

	fnsapi.RegisterRoutes(apiRouter)

	ens := &endpoints.Endpoints{
		DB:          db.NewPrefixedStore("/serverless-gateway/endpoints", kv),
		Logger:      log,
		FunctionsDB: fnsDB,
	}
	ensapi := &endpoints.HTTPAPI{Endpoints: ens}
	ensapi.RegisterRoutes(apiRouter)

	ps := &pubsub.PubSub{
		TopicsDB:        db.NewPrefixedStore("/serverless-gateway/topics", kv),
		SubscriptionsDB: db.NewPrefixedStore("/serverless-gateway/subscriptions", kv),
		PublishersDB:    db.NewPrefixedStore("/serverless-gateway/publishers", kv),
		FunctionsDB:     fnsDB,
		Logger:          log,
	}
	psapi := &pubsub.HTTPAPI{PubSub: ps}
	psapi.RegisterRoutes(apiRouter)

	return httptest.NewServer(apiRouter)
}
