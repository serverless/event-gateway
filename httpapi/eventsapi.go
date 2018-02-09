package httpapi

import (
	"net/http"
	"strconv"
	"time"
)

// StartEventsAPI creates a new gateway endpoint and listens for requests.
func StartEventsAPI(router http.Handler, config ServerConfig) {
	handler := &http.Server{
		Addr:         ":" + strconv.Itoa(int(config.Port)),
		Handler:      router,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 540 * time.Second,
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
