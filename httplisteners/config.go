package httplisteners

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
)

// StartConfigAPI creates a new configuration API server and listens for requests.
func StartConfigAPI(conf Config) {
	apiRouter := httprouter.New()

	fnsDB := db.NewPrefixedStore("/serverless-event-gateway/functions", conf.KV)
	fns := &functions.Functions{
		DB:     fnsDB,
		Logger: conf.Log,
	}
	fnsapi := &functions.HTTPAPI{Functions: fns}
	fnsapi.RegisterRoutes(apiRouter)

	ps := &pubsub.PubSub{
		TopicsDB:        db.NewPrefixedStore("/serverless-event-gateway/topics", conf.KV),
		SubscriptionsDB: db.NewPrefixedStore("/serverless-event-gateway/subscriptions", conf.KV),
		EndpointsDB:     db.NewPrefixedStore("/serverless-event-gateway/endpoints", conf.KV),
		FunctionsDB:     fnsDB,
		Logger:          conf.Log,
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
		Addr:         ":" + strconv.Itoa(int(conf.Port)),
		Handler:      apiHandler,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	h := handler{
		Conf:        conf,
		HTTPHandler: ev,
	}

	go func() {
		conf.ShutdownGuard.Add(1)
		h.listen()
		conf.ShutdownGuard.Done()
	}()
}
