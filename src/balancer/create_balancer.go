package balancer

import "math/rand"

func CreateBalancerFromConfig(config Config) Balancer {
	switch config.Balancer {
	case "least-loaded":
		return new(LeastLoadedBalancer)
	case "pkg-aware":
		return new(PkgAwareBalancer)
	case "random":
		rand.Seed(time.Now().Unix()) // For rand future calls
		return new(RandomBalancer)
	case "round-robin":
		return new(RoundRobinBalancer)
	}

	panic("Unknown balancer: " + config.Balancer)
}
