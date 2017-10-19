package httpapi

import (
	"context"
	"crypto/tls"
	"net/http"

	"go.uber.org/zap"
)

// Server is a context-aware http server.
type Server struct {
	Config
	HTTPHandler *http.Server
}

// Listen sets up a graceful shutdown mechanism and runs the http.Server.
func (s Server) Listen() {
	go func() {
		<-s.Config.ShutdownGuard.ShuttingDown
		s.HTTPHandler.Shutdown(context.Background())
	}()

	var err error

	if *s.Config.TLSCrt != "" && *s.Config.TLSKey != "" {
		s.HTTPHandler.TLSConfig = tlsConf
		s.HTTPHandler.TLSNextProto = map[string]func(*http.Server, *tls.Conn, http.Handler){}

		err = s.HTTPHandler.ListenAndServeTLS(*s.Config.TLSCrt, *s.Config.TLSKey)
	} else {
		err = s.HTTPHandler.ListenAndServe()
	}
	s.Config.Log.Error("HTTP server failed.", zap.Error(err))

	s.Config.ShutdownGuard.InitiateShutdown()
}
