package config

import (
	"github.com/disel-espol/olscheduler/balancer"
	"github.com/disel-espol/olscheduler/worker"
)

// Config holds then configured values and objects to be used by the scheduler.
type Config struct {
	Host          string
	Port          int
	LoadThreshold int
	Balancer      balancer.Balancer
	Registry      map[string][]string
	Workers       []*worker.Worker
	ReverseProxy  worker.ReverseProxy
}

func CreateDefaultConfig() Config {
	return Config{
		Host:          "localhost",
		Port:          9080,
		LoadThreshold: 3,
		Balancer:      new(balancer.RoundRobinBalancer),
		Registry:      make(map[string][]string),
		Workers:       make([]*worker.Worker, 0),
		ReverseProxy:  worker.NewHTTPReverseProxy(),
	}
}
