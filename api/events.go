package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/rs/cors"
	"github.com/serverless/event-gateway/internal/httpapi"
	"github.com/serverless/event-gateway/metrics"
	"github.com/serverless/event-gateway/router"
	"github.com/serverless/event-gateway/targetcache"
)

// StartEventsAPI creates a new gateway endpoint and listens for requests.
func StartEventsAPI(config httpapi.Config) httpapi.Server {
	targetCache := targetcache.New("/serverless-event-gateway", config.KV, config.Log)
	router := router.New(targetCache, metrics.DroppedPubSubEvents, config.Log)
	router.StartWorkers()

	handler := &http.Server{
		Addr:         ":" + strconv.Itoa(int(config.Port)),
		Handler:      cors.Default().Handler(router),
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
		router.Drain()
		config.ShutdownGuard.Done()
	}()

	return server
}
