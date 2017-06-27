package httplisteners

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/docker/libkv/store"
	"go.uber.org/zap"

	"github.com/serverless/event-gateway/util"
)

var (
	tlsConf = &tls.Config{
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
)

// Config contains information for an http listener to
// interact with its environment.
type Config struct {
	KV            store.Store
	Log           *zap.Logger
	TLSCrt        *string
	TLSKey        *string
	Port          uint
	ShutdownGuard *util.ShutdownGuard
}

// handler is a context-aware http server.
type handler struct {
	Conf        Config
	HTTPHandler *http.Server
}

// listen sets up a graceful shutdown mechanism and
// runs the http.Server.
func (h handler) listen() {
	go func() {
		<-h.Conf.ShutdownGuard.ShuttingDown
		h.HTTPHandler.Shutdown(context.Background())
	}()

	var err error
	if *h.Conf.TLSCrt != "" && *h.Conf.TLSKey != "" {
		h.HTTPHandler.TLSConfig = tlsConf
		h.HTTPHandler.TLSNextProto = map[string]func(*http.Server, *tls.Conn, http.Handler){}

		err = h.HTTPHandler.ListenAndServeTLS(*h.Conf.TLSCrt, *h.Conf.TLSKey)
	} else {
		err = h.HTTPHandler.ListenAndServe()
	}
	h.Conf.Log.Error("http server failed", zap.Error(err))

	h.Conf.ShutdownGuard.InitiateShutdown()
}
