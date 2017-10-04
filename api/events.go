package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/rs/cors"
	"github.com/serverless/event-gateway/internal/httpapi"
)

type EventsAPIConfig struct {
	httpapi.Config
	Router http.Handler
}

// StartEventsAPI creates a new gateway endpoint and listens for requests.
func StartEventsAPI(config EventsAPIConfig) httpapi.Server {
	handler := &http.Server{
		Addr:         ":" + strconv.Itoa(int(config.Port)),
		Handler:      cors.AllowAll().Handler(config.Router),
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	server := httpapi.Server{
		Config:      config.Config,
		HTTPHandler: handler,
	}

	config.ShutdownGuard.Add(1)
	go func() {
		server.Listen()
		config.ShutdownGuard.Done()
	}()

	return server
}
