package main

import (
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
	"github.com/serverless/event-gateway/httpapi"
	"github.com/serverless/event-gateway/metrics"
	"github.com/serverless/event-gateway/pubsub"
	"github.com/serverless/event-gateway/router"
	"github.com/serverless/event-gateway/targetcache"
	"github.com/serverless/event-gateway/util"
)

func init() {
	etcd.Register()
}

func startGateway(conf httpapi.HandlerConf) {
	targetCache := targetcache.New("/serverless-gateway", conf.KV, conf.Log)
	router := router.New(targetCache, metrics.DroppedPubSubEvents, conf.Log)
	ev := &http.Server{
		Addr:         ":" + strconv.Itoa(int(conf.Port)),
		Handler:      router,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	h := httpapi.Handler{
		Conf:        conf,
		HTTPHandler: ev,
	}

	go func() {
		conf.ShutdownGuard.Add(1)
		h.Listen()
		router.Drain()
		conf.ShutdownGuard.Done()
	}()
}

func startAPI(conf httpapi.HandlerConf) {
	apiRouter := httprouter.New()

	fnsDB := db.NewPrefixedStore("/serverless-gateway/functions", conf.KV)
	fns := &functions.Functions{
		DB:     fnsDB,
		Logger: conf.Log,
	}
	fnsapi := &functions.HTTPAPI{Functions: fns}
	fnsapi.RegisterRoutes(apiRouter)

	ens := &endpoints.Endpoints{
		DB:          db.NewPrefixedStore("/serverless-gateway/endpoints", conf.KV),
		Logger:      conf.Log,
		FunctionsDB: fnsDB,
	}
	ensapi := &endpoints.HTTPAPI{Endpoints: ens}
	ensapi.RegisterRoutes(apiRouter)

	ps := &pubsub.PubSub{
		TopicsDB:        db.NewPrefixedStore("/serverless-gateway/topics", conf.KV),
		SubscriptionsDB: db.NewPrefixedStore("/serverless-gateway/subscriptions", conf.KV),
		PublishersDB:    db.NewPrefixedStore("/serverless-gateway/publishers", conf.KV),
		FunctionsDB:     fnsDB,
		Logger:          conf.Log,
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
		Addr:         ":" + strconv.Itoa(int(conf.Port)),
		Handler:      handler,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	h := httpapi.Handler{
		Conf:        conf,
		HTTPHandler: ev,
	}

	go func() {
		conf.ShutdownGuard.Add(1)
		h.Listen()
		conf.ShutdownGuard.Done()
	}()
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

	dbHostStrings := strings.Split(*dbHosts, ",")

	prometheus.MustRegister(metrics.RequestDuration)
	prometheus.MustRegister(metrics.DroppedPubSubEvents)

	logCfg := zap.NewDevelopmentConfig()
	if !*verbose {
		logCfg = zap.NewProductionConfig()
		logCfg.DisableStacktrace = true
	}

	log, err := logCfg.Build()
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	shutdownGuard := util.NewShutdownGuard()

	if *embedMaster {
		db.EmbedEtcd(*embedDataDir, *embedPeerAddr, *embedCliAddr, shutdownGuard, *verbose)
	}

	kv, err := libkv.NewStore(
		store.ETCD,
		dbHostStrings,
		&store.Config{
			ConnectionTimeout: 10 * time.Second,
		},
	)
	if err != nil {
		log.Fatal("Cannot create kv client.", zap.Error(err))
	}

	// start API handler
	startAPI(httpapi.HandlerConf{
		KV:            kv,
		Log:           log,
		TLSCrt:        apiTLSCrt,
		TLSKey:        apiTLSKey,
		Port:          *apiPort,
		ShutdownGuard: shutdownGuard,
	})

	// start Event Gateway handler
	startGateway(httpapi.HandlerConf{
		KV:            kv,
		Log:           log,
		TLSCrt:        gatewayTLSCrt,
		TLSKey:        gatewayTLSKey,
		Port:          *gatewayPort,
		ShutdownGuard: shutdownGuard,
	})

	shutdownGuard.Wait()
}
