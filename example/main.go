package main

import (
	"github.com/goroute/route"
	"github.com/goroute/static"
	"log"
	"net/http"
)

func main() {
	mux := route.NewServeMux()

	mux.Use(static.New(
		static.Browse(true),
		static.Root("../testdata/browse"),
	))

	log.Fatal(http.ListenAndServe(":9000", mux))
}
