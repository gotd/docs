//go:build !js

// Command serve is a tiny static file server for the WASM demo.
//
// It exists so the example is runnable with the Go toolchain alone, without a
// separate web server. Run it from the example directory after building
// main.wasm (see the Makefile):
//
//	go run ./serve
package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "listen address")
	dir := flag.String("dir", ".", "directory to serve")
	flag.Parse()

	log.Printf("serving %s on http://%s", *dir, *addr)
	if err := http.ListenAndServe(*addr, http.FileServer(http.Dir(*dir))); err != nil {
		log.Fatal(err)
	}
}
