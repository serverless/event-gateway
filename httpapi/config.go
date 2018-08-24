package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/subscription"
	"github.com/serverless/event-gateway/subscription/cors"
)

// StartConfigAPI creates a new configuration API server and listens for requests.
func StartConfigAPI(eventtypes event.Service, functions function.Service, subscriptions subscription.Service, corses cors.Service, config ServerConfig) {
	router := httprouter.New()
	api := &HTTPAPI{
		EventTypes:    eventtypes,
		Functions:     functions,
		Subscriptions: subscriptions,
		CORSes:        corses,
	}
	api.RegisterRoutes(router)

	handler := &http.Server{
		Addr:         ":" + strconv.Itoa(int(config.Port)),
		Handler:      metricsReporter{Handler: router},
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

type metricsReporter struct {
	Handler http.Handler
}

func (m metricsReporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	m.Handler.ServeHTTP(w, r)
	metricConfigRequestDuration.Observe(time.Since(start).Seconds())
}
