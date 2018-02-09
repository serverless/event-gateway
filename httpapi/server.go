package httpapi

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/serverless/event-gateway/internal/sync"
	"go.uber.org/zap"
)

// ServerConfig contains information for an HTTP listener to interact with its environment.
type ServerConfig struct {
	Log           *zap.Logger
	TLSCrt        *string
	TLSKey        *string
	Port          uint
	ShutdownGuard *sync.ShutdownGuard
}

var tlsConf = &tls.Config{
	MinVersion:               tls.VersionTLS12,
	CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
	PreferServerCipherSuites: true,
	CipherSuites: []uint16{
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	},
}

// Server is a context-aware http server.
type Server struct {
	Config      ServerConfig
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
