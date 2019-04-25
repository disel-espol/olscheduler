package config

import "github.com/disel-espol/olscheduler/balancer"

func createBalancerFromConfig(c JSONConfig) balancer.Balancer {
	switch c.Balancer {
	case "least-loaded":
		return balancer.NewLeastLoadedFromJSONSlice(c.Workers)
	case "pkg-aware":
		return balancer.NewPackageAwareFromJSONSlice(c.Workers, uint(c.LoadThreshold))
	case "random":
		return balancer.NewRandomFromJSONSlice(c.Workers)
	case "round-robin":
		return balancer.NewRoundRobinFromJSONSlice(c.Workers)
	case "hashing-bounded":
		return balancer.NewConsistentHashingBoundedFromJSONSlice(c.Workers)
	}

	panic("Unknown balancer: " + c.Balancer)
}
