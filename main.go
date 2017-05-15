package main

import (
	"fmt"
	"net/http"
)

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello from gateway!")
}

func main() {
	http.HandleFunc("/status", hello)
	http.HandleFunc("/v0/gateway", hello)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
