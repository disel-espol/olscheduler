package config

import (
	"../balancer"
	"../worker"
)

// Config holds then configured values and objects to be used by the scheduler.
type Config struct {
	Host          string
	Port          int
	LoadThreshold int
	Balancer      balancer.Balancer
	Registry      map[string][]string
	Workers       []*worker.Worker
	ReverseProxy  *worker.ReverseProxy
}
