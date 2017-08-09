package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/rs/cors"
	"github.com/serverless/event-gateway/metrics"
	"github.com/serverless/event-gateway/router"
	"github.com/serverless/event-gateway/targetcache"
	"github.com/serverless/event-gateway/util/httpapi"
)

// StartEventsAPI creates a new gateway endpoint and listens for requests.
func StartEventsAPI(config httpapi.Config) {
	targetCache := targetcache.New("/serverless-event-gateway", config.KV, config.Log)
	router := router.New(targetCache, metrics.DroppedPubSubEvents, config.Log)
	router.StartWorkers()
	ev := &http.Server{
		Addr:         ":" + strconv.Itoa(int(config.Port)),
		Handler:      cors.Default().Handler(router),
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
		router.Drain()
		config.ShutdownGuard.Done()
	}()
}
