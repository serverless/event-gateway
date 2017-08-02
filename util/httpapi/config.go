package httpapi

import (
	"crypto/tls"

	"github.com/docker/libkv/store"
	"github.com/serverless/event-gateway/util"
	"go.uber.org/zap"
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
