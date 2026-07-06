package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
    // 1. Read a port from a flag so you can start several: -port=9001, 9002, ...
    port := flag.String("port", "9001", "port to listen on")
    flag.Parse()

    // 2. A handler that echoes which backend answered + the path
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello from backend on port %s (path: %s)\n", *port, r.URL.Path)
    })

    // 3. A health endpoint for later (your health checker will hit this)
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })

    // 4. Start listening
    addr := ":" + *port
    log.Printf("backend listening on %s", addr)
    log.Fatal(http.ListenAndServe(addr, nil))
}
