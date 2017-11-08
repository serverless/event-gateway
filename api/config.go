package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/internal/httpapi"
	"github.com/serverless/event-gateway/internal/kv"
	"github.com/serverless/event-gateway/subscriptions"
)

// StartConfigAPI creates a new configuration API server and listens for requests.
func StartConfigAPI(config httpapi.Config) {
	router := httprouter.New()

	functionsDB := kv.NewPrefixedStore("/serverless-event-gateway/functions", config.KV)
	functionService := &functions.Functions{
		DB:  functionsDB,
		Log: config.Log,
	}
	functionsAPI := &functions.HTTPAPI{Functions: functionService}
	functionsAPI.RegisterRoutes(router)

	subscriptionsService := &subscriptions.Subscriptions{
		SubscriptionsDB: kv.NewPrefixedStore("/serverless-event-gateway/subscriptions", config.KV),
		EndpointsDB:     kv.NewPrefixedStore("/serverless-event-gateway/endpoints", config.KV),
		FunctionsDB:     functionsDB,
		Log:             config.Log,
	}
	subscriptionsAPI := &subscriptions.HTTPAPI{Subscriptions: subscriptionsService}
	subscriptionsAPI.RegisterRoutes(router)

	router.GET("/v1/status", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	router.Handler("GET", "/metrics", promhttp.Handler())

	handler := &http.Server{
		Addr:         ":" + strconv.Itoa(int(config.Port)),
		Handler:      metricsReporter{router},
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	server := httpapi.Server{
		Config:      config,
		HTTPHandler: handler,
	}

	config.ShutdownGuard.Add(1)
	go func() {
		server.Listen()
		config.ShutdownGuard.Done()
	}()
}
