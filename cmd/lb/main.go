package main

import (
	"log"

	"github.com/neautrino/loadbalancer/internal"
	"github.com/neautrino/loadbalancer/internal/algorithms"
)

func main() {
    lb, err := internal.NewLoadBalancer(":8080", []string{
        "http://localhost:9001",
        "http://localhost:9002",
        "http://localhost:9003",
    }, &algorithms.LeastConn{})
    if err != nil {
        log.Fatal(err)
    }

    log.Fatal(lb.Start())
}
