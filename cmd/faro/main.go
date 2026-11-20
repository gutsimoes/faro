package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gutsimoes/faro/server"
)

const version = "0.0-dev"

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	flag.Parse()

	log.Printf("faro %s listening on %s", version, *addr)
	log.Fatal(http.ListenAndServe(*addr, server.NewHandler()))
}
