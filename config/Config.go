package config

import (
	"net/url"

	"github.com/disel-espol/olscheduler/balancer"
	"github.com/disel-espol/olscheduler/proxy"
)

// Config holds then configured values and objects to be used by the scheduler.
type Config struct {
	Host         string
	Port         int
	Balancer     balancer.Balancer
	Registry     map[string][]string
	ReverseProxy proxy.ReverseProxy
}

func CreateDefaultConfig() Config {
	return Config{
		Host:         "localhost",
		Port:         9080,
		Balancer:     balancer.NewRoundRobin(make([]url.URL, 0)),
		Registry:     make(map[string][]string),
		ReverseProxy: proxy.NewHTTPReverseProxy(),
	}
}
