package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/serverless/event-gateway/api"
)

func init() {
	prometheus.MustRegister(requestDuration)
}

// StartConfigAPI creates a new configuration API server and listens for requests.
func StartConfigAPI(functions api.FunctionService, subscriptions api.SubscriptionService, config ServerConfig) {
	router := httprouter.New()
	api := &HTTPAPI{
		Functions:     functions,
		Subscriptions: subscriptions,
	}
	api.RegisterRoutes(router)

	handler := &http.Server{
		Addr:         ":" + strconv.Itoa(int(config.Port)),
		Handler:      metricsReporter{router},
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	server := Server{
		Config:      config,
		HTTPHandler: handler,
	}

	config.ShutdownGuard.Add(1)
	go func() {
		server.Listen()
		config.ShutdownGuard.Done()
	}()
}

var requestDuration = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Namespace: "gateway",
		Subsystem: "config",
		Name:      "request_duration_seconds",
		Help:      "Bucketed histogram of request duration of config API requests",
		Buckets:   prometheus.ExponentialBuckets(0.0005, 2, 16),
	})

type metricsReporter struct {
	Handler http.Handler
}

func (m metricsReporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	m.Handler.ServeHTTP(w, r)
	requestDuration.Observe(time.Since(start).Seconds())
}
