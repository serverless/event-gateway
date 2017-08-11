package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/rs/cors"
	"github.com/serverless/event-gateway/internal/cache"
	"github.com/serverless/event-gateway/internal/httpapi"
	"github.com/serverless/event-gateway/internal/metrics"
	"github.com/serverless/event-gateway/router"
)

// StartEventsAPI creates a new gateway endpoint and listens for requests.
func StartEventsAPI(config httpapi.Config) httpapi.Server {
	targetCache := cache.NewTarget("/serverless-event-gateway", config.KV, config.Log)
	router := router.New(targetCache, metrics.DroppedPubSubEvents, config.Log)
	router.StartWorkers()

	handler := &http.Server{
		Addr:         ":" + strconv.Itoa(int(config.Port)),
		Handler:      cors.AllowAll().Handler(router),
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
