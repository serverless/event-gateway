package httpapi

import (
	"context"
	"crypto/tls"
	"net/http"

	"go.uber.org/zap"
)

// Handler is a context-aware http server.
type Handler struct {
	Config
	HTTPHandler *http.Server
}

// Listen sets up a graceful shutdown mechanism and runs the http.Server.
func (h Handler) Listen() {
	go func() {
		<-h.Config.ShutdownGuard.ShuttingDown
		h.HTTPHandler.Shutdown(context.Background())
	}()

	var err error
	if *h.Config.TLSCrt != "" && *h.Config.TLSKey != "" {
		h.HTTPHandler.TLSConfig = tlsConf
		h.HTTPHandler.TLSNextProto = map[string]func(*http.Server, *tls.Conn, http.Handler){}

		err = h.HTTPHandler.ListenAndServeTLS(*h.Config.TLSCrt, *h.Config.TLSKey)
	} else {
		err = h.HTTPHandler.ListenAndServe()
	}
	h.Config.Log.Error("http server failed", zap.Error(err))

	h.Config.ShutdownGuard.InitiateShutdown()
}
