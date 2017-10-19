package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/serverless/event-gateway/internal/httpapi"
)

// StartEventsAPI creates a new gateway endpoint and listens for requests.
func StartEventsAPI(config httpapi.Config, router http.Handler) {
	handler := &http.Server{
		Addr:         ":" + strconv.Itoa(int(config.Port)),
		Handler:      router,
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
