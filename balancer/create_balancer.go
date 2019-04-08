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
		pkg := new(PkgAwareBalancer);
		pkg.init(config.Workers, config.LoadThreshold);
		return pkg;
	case "random":
		rand.Seed(time.Now().Unix()) // For rand future calls
		return new(RandomBalancer)
	case "round-robin":
		return new(RoundRobinBalancer)
	case "hashing-bounded":
		c := new(ConsistentHashingBounded)
		c.init(config.Workers)
		return c
	}

	panic("Unknown balancer: " + config.Balancer)
}
