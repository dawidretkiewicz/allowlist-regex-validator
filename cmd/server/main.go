package main

import (
	"log"
	"net/http"
	"os"

	"regex-validator/internal/api"
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	mux := http.NewServeMux()
	h := api.NewHandler()
	h.Register(mux)

	mux.Handle("/", http.FileServer(http.Dir("web/static")))

	log.Printf("regex validator server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
