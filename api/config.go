package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/serverless/event-gateway/db"
	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/metrics"
	"github.com/serverless/event-gateway/pubsub"
	"github.com/serverless/event-gateway/util/httpapi"
)

// StartConfigAPI creates a new configuration API server and listens for requests.
func StartConfigAPI(config httpapi.Config) {
	apiRouter := httprouter.New()

	fnsDB := db.NewPrefixedStore("/serverless-event-gateway/functions", config.KV)
	fns := &functions.Functions{
		DB:  fnsDB,
		Log: config.Log,
	}
	fnsapi := &functions.HTTPAPI{Functions: fns}
	fnsapi.RegisterRoutes(apiRouter)

	ps := &pubsub.PubSub{
		TopicsDB:        db.NewPrefixedStore("/serverless-event-gateway/topics", config.KV),
		SubscriptionsDB: db.NewPrefixedStore("/serverless-event-gateway/subscriptions", config.KV),
		EndpointsDB:     db.NewPrefixedStore("/serverless-event-gateway/endpoints", config.KV),
		FunctionsDB:     fnsDB,
		Log:             config.Log,
	}
	psapi := &pubsub.HTTPAPI{PubSub: ps}
	psapi.RegisterRoutes(apiRouter)

	apiRouter.GET("/v1/status", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	apiRouter.Handler("GET", "/metrics", prometheus.Handler())

	apiHandler := metrics.HTTPLogger{
		Handler:         apiRouter,
		RequestDuration: metrics.RequestDuration,
	}
	ev := &http.Server{
		Addr:         ":" + strconv.Itoa(int(config.Port)),
		Handler:      apiHandler,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	h := httpapi.Handler{
		Config:      config,
		HTTPHandler: ev,
	}

	go func() {
		config.ShutdownGuard.Add(1)
		h.Listen()
		config.ShutdownGuard.Done()
	}()
}
