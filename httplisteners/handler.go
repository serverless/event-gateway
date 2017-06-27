package httplisteners

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/docker/libkv/store"
	"go.uber.org/zap"

	"github.com/serverless/event-gateway/util"
)

// HandlerConf contains information for an http listener to
// interact with its environment.
type HandlerConf struct {
	KV            store.Store
	Log           *zap.Logger
	TLSCrt        *string
	TLSKey        *string
	Port          uint
	ShutdownGuard *util.ShutdownGuard
}

// Handler is a context-aware http server.
type Handler struct {
	Conf        HandlerConf
	HTTPHandler *http.Server
}

// Listen sets up a graceful shutdown mechanism and
// runs the http.Server.
func (h Handler) Listen() {
	go func() {
		<-h.Conf.ShutdownGuard.ShuttingDown
		h.HTTPHandler.Shutdown(context.Background())
	}()

	var err error
	if *h.Conf.TLSCrt != "" && *h.Conf.TLSKey != "" {
		h.HTTPHandler.TLSConfig = tlsConf()
		h.HTTPHandler.TLSNextProto = map[string]func(*http.Server, *tls.Conn, http.Handler){}

		err = h.HTTPHandler.ListenAndServeTLS(*h.Conf.TLSCrt, *h.Conf.TLSKey)
	} else {
		err = h.HTTPHandler.ListenAndServe()
	}
	h.Conf.Log.Error("http server failed", zap.Error(err))

	h.Conf.ShutdownGuard.InitiateShutdown()
}

func tlsConf() *tls.Config {
	return &tls.Config{
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
}
