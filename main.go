package main

import (
	"flag"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/julienschmidt/httprouter"

	"github.com/serverless/gateway/db"
	"github.com/serverless/gateway/endpoints"
	"github.com/serverless/gateway/functions"
	"github.com/serverless/gateway/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	verbose := flag.Bool("verbose", false, "Verbose logging.")
	dbHosts := flag.String("db-hosts", "localhost:2379", "Comma-separated list of database hosts to connect to.")
	embedMaster := flag.Bool("embed-master", false, "Run embedded etcd for testing.")
	embedPeerAddr := flag.String("embed-peer-addr", "http://localhost:2380", "Address for testing embedded etcd to receive peer connections.")
	embedCliAddr := flag.String("embed-cli-addr", "http://localhost:2379", "Address for testing embedded etcd to receive client connections.")
	embedDataDir := flag.String("embed-data-dir", "default.etcd", "Path for testing embedded etcd to store its state.")
	flag.Parse()

	prometheus.MustRegister(metrics.DurationMetric)

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
			*embedCliAddr, shutdownInitiateChan, logger, *verbose)
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

	router := httprouter.New()

	fdb := db.NewReactiveCfgStore("/serverless-gateway/functions", dbHostStrings, logger)
	fns := &functions.Functions{
		DB:     fdb,
		Logger: logger,
	}
	fdb.React(fns, shutdownInitiateChan)
	fnsapi := &functions.HTTPAPI{Functions: fns}
	fnsapi.RegisterRoutes(router)

	edb := db.NewReactiveCfgStore("/serverless-gateway/endpoints", dbHostStrings, logger)
	ens := &endpoints.Endpoints{
		DB:      edb,
		Invoker: fns,
		Logger:  logger,
	}
	edb.React(ens, shutdownInitiateChan)
	ensapi := &endpoints.HTTPAPI{Endpoints: ens}
	ensapi.RegisterRoutes(router)

	router.GET("/status", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	router.Handler("GET", "/v0/gateway/metrics", prometheus.Handler())
	err = http.ListenAndServe(":8080", metrics.HTTPLogger{router, metrics.DurationMetric})
	logger.Error("server failed", zap.Error(err))
	close(shutdownInitiateChan)
	<-shutdownCompleteChan
}
