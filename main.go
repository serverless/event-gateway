package main

import (
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/julienschmidt/httprouter"
	"github.com/serverless/gateway/db"
	"github.com/serverless/gateway/functions"
)

func main() {
	db, err := db.New()
	if err != nil {
		log.Printf("loading db file failed: %q", err)
		return
	}
	defer db.Close()

	router := httprouter.New()
	router.GET("/status", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})

	fns := &functions.Functions{
		DB:        db,
		AWSLambda: lambda.New(session.New(aws.NewConfig())),
	}
	fnsapi := &functions.HTTPAPI{Functions: fns}
	fnsapi.RegisterRoutes(router)

	log.Fatal(http.ListenAndServe(":8080", router))
}
