package tests

import (
	"net/http/httptest"

	"github.com/docker/libkv/store"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"

	"github.com/serverless/event-gateway/db"
	"github.com/serverless/event-gateway/endpoints"
	"github.com/serverless/event-gateway/functions"
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

	return httptest.NewServer(apiRouter)
}
