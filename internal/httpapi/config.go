package httpapi

import (
	"crypto/tls"

	"github.com/serverless/event-gateway/internal/sync"
	"github.com/serverless/libkv/store"
	"go.uber.org/zap"
)

// Config contains information for an http listener to interact with its environment.
type Config struct {
	KV            store.Store
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
