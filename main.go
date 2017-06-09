package main

import (
	"flag"
	"net/http"

	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/julienschmidt/httprouter"
	"github.com/serverless/gateway/db"
	"github.com/serverless/gateway/endpoints"
	"github.com/serverless/gateway/functions"

	"github.com/prometheus/client_golang/prometheus"
)

var durationMetric = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Name:    "gateway_request_duration_milliseconds",
		Help:    "Request duration distribution",
		Buckets: prometheus.ExponentialBuckets(0.1, 1.1, 140),
	})

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	prometheus.MustRegister(durationMetric)

	dataDir := flag.String("data-dir", "", "Path to a data directory to store instance state.")
	flag.Parse()

	db, err := db.New(*dataDir)
	if err != nil {
		logger.Info("loading db file failed", zap.Error(err))
		return
	}
	defer db.Close()

	router := httprouter.New()

	fns := &functions.Functions{
		DB:        db,
		AWSLambda: lambda.New(session.New(aws.NewConfig())),
		Logger:    logger,
	}
	fnsapi := &functions.HTTPAPI{Functions: fns}
	fnsapi.RegisterRoutes(router)

	ens := &endpoints.Endpoints{
		DB:      db,
		Invoker: fns,
		Logger:  logger,
	}
	ensapi := &endpoints.HTTPAPI{Endpoints: ens}
	ensapi.RegisterRoutes(router)

	router.GET("/status", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	router.Handler("GET", "/metrics", prometheus.Handler())
	err = http.ListenAndServe(":8080", HTTPLogger{router, durationMetric})
	logger.Fatal("server failed", zap.Error(err))
}
