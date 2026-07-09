package main

import (
	"fmt"
	"log"

	"github.com/neautrino/loadbalancer/internal"
	"github.com/neautrino/loadbalancer/internal/algorithms"
	"github.com/neautrino/loadbalancer/internal/config"
)

func BuildStrategy(name string, replicas int) (algorithms.Strategy, error) {
	switch name {
	case "round_robin":     return algorithms.NewRoundRobin(), nil
	case "least_conn":      return algorithms.NewLeastConn(), nil
	case "weighted":        return algorithms.NewWeightedRoundRobin(), nil
	case "ip_hash":         return algorithms.NewIpHash(), nil
	case "consistent_hash": return algorithms.NewConsistentHash(replicas), nil
	default:                return nil, fmt.Errorf("unknown algorithm: %q", name)
	}
}

func main() {
    cfg, err := config.Load("config.yaml")
    if err != nil {
        log.Fatal(err)
    }

    strategy, err := BuildStrategy(cfg.Algorithm, 100)
    if err != nil {
        log.Fatal(err)
    }

    lb, err := internal.NewLoadBalancer(cfg , strategy)
    if err != nil {
        log.Fatal(err)
    }

    log.Fatal(lb.Start())
}
