package main

import (
	"flag"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/julienschmidt/httprouter"

	"github.com/serverless/gateway/db"
	"github.com/serverless/gateway/endpoints"
	"github.com/serverless/gateway/functions"
)

func main() {
	verbose := flag.Bool("verbose", false, "Verbose logging.")
	dbType := flag.String("db-type", "etcd", "Kind of backing database. One of etcd, zookeeper, or consul.")
	dbHosts := flag.String("db-hosts", "localhost:2379", "Comma-separated list of database hosts to connect to.")
	embedMaster := flag.Bool("embed-master", false, "Run embedded etcd for testing.")
	embedPeerAddr := flag.String("embed-peer-addr", "http://localhost:2380", "Address for testing embedded etcd to receive peer connections.")
	embedCliAddr := flag.String("embed-cli-addr", "http://localhost:2379", "Address for testing embedded etcd to receive client connections.")
	embedDataDir := flag.String("embed-data-dir", "default.etcd", "Path for testing embedded etcd to store its state.")
	flag.Parse()

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
	router.GET("/status", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})

	fdb := db.NewReactiveCfgStore("/serverless-gateway/functions", *dbType, dbHostStrings, logger)
	fns := &functions.Functions{
		DB:        fdb,
		AWSLambda: lambda.New(session.New(aws.NewConfig())),
		Logger:    logger,
	}
	fdb.React(fns, shutdownInitiateChan)
	fnsapi := &functions.HTTPAPI{Functions: fns}
	fnsapi.RegisterRoutes(router)

	edb := db.NewReactiveCfgStore("/serverless-gateway/endpoints", *dbType, dbHostStrings, logger)
	ens := &endpoints.Endpoints{
		DB:      edb,
		Invoker: fns,
		Logger:  logger,
	}
	edb.React(ens, shutdownInitiateChan)
	ensapi := &endpoints.HTTPAPI{Endpoints: ens}
	ensapi.RegisterRoutes(router)

	err = http.ListenAndServe(":8080", router)
	logger.Error("server failed, shutting down", zap.Error(err))

	close(shutdownInitiateChan)
	<-shutdownCompleteChan
}
