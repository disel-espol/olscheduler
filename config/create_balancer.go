package config

import (
	"math/rand"
	"time"

	"../balancer"
)

func createBalancerFromConfig(c JSONConfig) balancer.Balancer {
	switch c.Balancer {
	case "least-loaded":
		return new(balancer.LeastLoadedBalancer)
	case "pkg-aware":
		b := new(balancer.PkgAwareBalancer)
		b.Init(c.Workers, c.LoadThreshold)
		return b
	case "random":
		rand.Seed(time.Now().Unix()) // For rand future calls
		return new(balancer.RandomBalancer)
	case "round-robin":
		return new(balancer.RoundRobinBalancer)
	case "hashing-bounded":
		b := new(balancer.ConsistentHashingBounded)
		b.Init(c.Workers)
		return b
	}

	panic("Unknown balancer: " + c.Balancer)
}
