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
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dataDir := flag.String("data-dir", "", "Path to a data directory to store instance state.")
	flag.Parse()

	db, err := db.New(*dataDir)
	if err != nil {
		logger.Info("loading db file failed", zap.Error(err))
		return
	}
	defer db.Close()

	router := httprouter.New()
	router.GET("/status", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})

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

	err = http.ListenAndServe(":8080", router)
	logger.Fatal("server failed", zap.Error(err))
}
