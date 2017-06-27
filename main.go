package main

import (
	"crypto/tls"
	"flag"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/etcd"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/serverless/event-gateway/db"
	"github.com/serverless/event-gateway/endpoints"
	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/metrics"
	"github.com/serverless/event-gateway/pubsub"
	"github.com/serverless/event-gateway/router"
	"github.com/serverless/event-gateway/targetcache"
)

func init() {
	etcd.Register()
}

func main() {
	verbose := flag.Bool("verbose", false, "Verbose logging.")
	dbHosts := flag.String("db-hosts", "localhost:2379", "Comma-separated list of database hosts to connect to.")
	embedMaster := flag.Bool("embed-master", false, "Run embedded etcd for testing.")
	embedPeerAddr := flag.String("embed-peer-addr", "http://localhost:2380", "Address for testing embedded etcd to receive peer connections.")
	embedCliAddr := flag.String("embed-cli-addr", "http://localhost:2379", "Address for testing embedded etcd to receive client connections.")
	embedDataDir := flag.String("embed-data-dir", "default.etcd", "Path for testing embedded etcd to store its state.")
	apiPort := flag.Uint("api-port", 8081, "Port to serve configuration API on.")
	apiTLSCrt := flag.String("api-tls-cert", "", "Path to API TLS certificate file.")
	apiTLSKey := flag.String("api-tls-key", "", "Path to API TLS key file.")
	gatewayTLSCrt := flag.String("gateway-tls-cert", "", "Path to gateway TLS certificate file.")
	gatewayTLSKey := flag.String("gateway-tls-key", "", "Path to gateway TLS key file.")
	gatewayPort := flag.Uint("gateway-port", 8080, "Port to serve configured endpoints on.")
	flag.Parse()

	prometheus.MustRegister(metrics.RequestDuration)
	prometheus.MustRegister(metrics.DroppedPubSubEvents)

	dbHostStrings := strings.Split(*dbHosts, ",")

	cfg := zap.NewDevelopmentConfig()
	if !*verbose {
		cfg = zap.NewProductionConfig()
		cfg.DisableStacktrace = true
	}

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	shutdownInitiateChan := make(chan struct{})
	shutdownCompleteChan := make(chan struct{})
	if *embedMaster {
		startedChan, stoppedChan := db.EmbedEtcd(*embedDataDir, *embedPeerAddr,
			*embedCliAddr, shutdownInitiateChan, *verbose)
		select {
		case <-startedChan:
			defer func() {
				<-stoppedChan
				close(shutdownCompleteChan)
			}()
		case <-stoppedChan:
			logger.Fatal("Failed to start embedded etcd.")
		}
	}

	kv, err := libkv.NewStore(
		store.ETCD,
		dbHostStrings,
		&store.Config{
			ConnectionTimeout: 10 * time.Second,
		},
	)

	if err != nil {
		logger.Fatal("Cannot create kv client.",
			zap.Error(err))
	}

	// start API handler
	go func() {
		apiRouter := httprouter.New()

		fnsDB := db.NewPrefixedStore("/serverless-gateway/functions", kv)
		fns := &functions.Functions{
			DB:     fnsDB,
			Logger: logger,
		}
		fnsapi := &functions.HTTPAPI{Functions: fns}
		fnsapi.RegisterRoutes(apiRouter)

		ens := &endpoints.Endpoints{
			DB:          db.NewPrefixedStore("/serverless-gateway/endpoints", kv),
			Logger:      logger,
			FunctionsDB: fnsDB,
		}
		ensapi := &endpoints.HTTPAPI{Endpoints: ens}
		ensapi.RegisterRoutes(apiRouter)

		ps := &pubsub.PubSub{
			TopicsDB:        db.NewPrefixedStore("/serverless-gateway/topics", kv),
			SubscriptionsDB: db.NewPrefixedStore("/serverless-gateway/subscriptions", kv),
			PublishersDB:    db.NewPrefixedStore("/serverless-gateway/publishers", kv),
			FunctionsDB:     fnsDB,
			Logger:          logger,
		}
		psapi := &pubsub.HTTPAPI{PubSub: ps}
		psapi.RegisterRoutes(apiRouter)

		apiRouter.GET("/status", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
		apiRouter.Handler("GET", "/v0/gateway/metrics", prometheus.Handler())

		handler := metrics.HTTPLogger{
			Handler:         apiRouter,
			RequestDuration: metrics.RequestDuration,
		}
		ev := &http.Server{
			Addr:         ":" + strconv.Itoa(int(*apiPort)),
			Handler:      handler,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		}

		if *apiTLSCrt != "" && *apiTLSKey != "" {
			cfg := &tls.Config{
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

			ev.TLSConfig = cfg
			ev.TLSNextProto = map[string]func(*http.Server, *tls.Conn, http.Handler){}
			err = ev.ListenAndServeTLS(*apiTLSCrt, *apiTLSKey)
		} else {
			err = ev.ListenAndServe()
		}

		logger.Error("api server failed", zap.Error(err))
		close(shutdownInitiateChan)
	}()

	// start Event Gateway handler
	go func() {
		targetCache := targetcache.New("/serverless-gateway", kv, logger)
		router := router.New(targetCache, metrics.DroppedPubSubEvents, logger)
		ev := &http.Server{
			Addr:         ":" + strconv.Itoa(int(*gatewayPort)),
			Handler:      router,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		}

		if *gatewayTLSCrt != "" && *gatewayTLSKey != "" {
			cfg := &tls.Config{
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

			ev.TLSConfig = cfg
			ev.TLSNextProto = map[string]func(*http.Server, *tls.Conn, http.Handler){}
			err = ev.ListenAndServeTLS(*gatewayTLSCrt, *gatewayTLSKey)
		} else {
			err = ev.ListenAndServe()
		}
		logger.Error("gateway server failed", zap.Error(err))
		close(shutdownInitiateChan)
		router.Drain()
	}()
	<-shutdownCompleteChan
}
