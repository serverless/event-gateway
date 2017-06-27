package httpapi

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/docker/libkv/store"
	"go.uber.org/zap"

	"github.com/serverless/event-gateway/util"
)

type HandlerConf struct {
	KV            store.Store
	Log           *zap.Logger
	TLSCrt        *string
	TLSKey        *string
	Port          uint
	ShutdownGuard *util.ShutdownGuard
}

type Handler struct {
	Conf        HandlerConf
	HTTPHandler *http.Server
}

func (h Handler) Listen() {
	go func() {
		<-h.Conf.ShutdownGuard.ShuttingDown
		h.HTTPHandler.Shutdown(context.Background())
	}()

	var err error
	if *h.Conf.TLSCrt != "" && *h.Conf.TLSKey != "" {
		h.HTTPHandler.TLSConfig = TLSConf()
		h.HTTPHandler.TLSNextProto = map[string]func(*http.Server, *tls.Conn, http.Handler){}

		err = h.HTTPHandler.ListenAndServeTLS(*h.Conf.TLSCrt, *h.Conf.TLSKey)
	} else {
		err = h.HTTPHandler.ListenAndServe()
	}
	h.Conf.Log.Error("http server failed", zap.Error(err))

	h.Conf.ShutdownGuard.InitiateShutdown()
}

func TLSConf() *tls.Config {
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
