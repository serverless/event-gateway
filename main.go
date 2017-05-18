package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/serverless/gateway/db"
	"github.com/serverless/gateway/functions"
)

func main() {
	db, err := db.New()
	if err != nil {
		log.Fatalf("loading db file failed: %q", err)
	}
	defer db.Close()

	router := httprouter.New()
	router.GET("/status", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})

	fns := &functions.Functions{DB: db}
	fnsapi := &functions.HTTPAPI{Functions: fns}
	fnsapi.RegisterRoutes(router)

	log.Fatal(http.ListenAndServe(":8080", router))
}
