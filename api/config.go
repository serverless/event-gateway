package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/cors"

	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/internal/httpapi"
	"github.com/serverless/event-gateway/internal/kv"
	"github.com/serverless/event-gateway/metrics"
	"github.com/serverless/event-gateway/pubsub"
)

// StartConfigAPI creates a new configuration API server and listens for requests.
func StartConfigAPI(config httpapi.Config) httpapi.Server {
	router := httprouter.New()

	functionsDB := kv.NewPrefixedStore("/serverless-event-gateway/functions", config.KV)
	functionService := &functions.Functions{
		DB:  functionsDB,
		Log: config.Log,
	}
	functionsAPI := &functions.HTTPAPI{Functions: functionService}
	functionsAPI.RegisterRoutes(router)

	pubsubService := &pubsub.PubSub{
		TopicsDB:        kv.NewPrefixedStore("/serverless-event-gateway/topics", config.KV),
		SubscriptionsDB: kv.NewPrefixedStore("/serverless-event-gateway/subscriptions", config.KV),
		EndpointsDB:     kv.NewPrefixedStore("/serverless-event-gateway/endpoints", config.KV),
		FunctionsDB:     functionsDB,
		Log:             config.Log,
	}
	pubsubAPI := &pubsub.HTTPAPI{PubSub: pubsubService}
	pubsubAPI.RegisterRoutes(router)

	router.GET("/v1/status", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	router.Handler("GET", "/metrics", prometheus.Handler())

	handler := &http.Server{
		Addr:         ":" + strconv.Itoa(int(config.Port)),
		Handler:      cors.Default().Handler(metrics.HTTPLogger{Handler: router, RequestDuration: metrics.RequestDuration}),
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	server := httpapi.Server{
		Config:      config,
		HTTPHandler: handler,
	}

	go func() {
		config.ShutdownGuard.Add(1)
		server.Listen()
		config.ShutdownGuard.Done()
	}()

	return server
}
