package balancer

import (
	"../schutil"
	"math/rand"
	"time"
)

func CreateBalancerFromConfig(config schutil.Config) Balancer {
	switch config.Balancer {
	case "least-loaded":
		return new(LeastLoadedBalancer)
	case "pkg-aware":
		return &PkgAwareBalancer{config.LoadThreshold}
	case "random":
		rand.Seed(time.Now().Unix()) // For rand future calls
		return new(RandomBalancer)
	case "round-robin":
		return new(RoundRobinBalancer)
	}

	panic("Unknown balancer: " + config.Balancer)
}
