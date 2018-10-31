package scheduler_test

import (
	"testing"

	"../balancer"
	"../scheduler"
	"../worker"
)

func newScheduler() *scheduler.Scheduler {
	return scheduler.NewScheduler(
		make(map[string][]string),
		new(balancer.RoundRobinBalancer),
		make([]*worker.Worker, 0))
}

func TestAddWorkers(t *testing.T) {
	s := newScheduler()

	s.AddWorkers([]string{"localhost:9001", "localhost:9002"})

	totalWorkers := s.GetTotalWorkers()
	if totalWorkers != 2 {
		t.Errorf("Expected 2 workers but got %d instead", totalWorkers)
	}
}

func TestRemoveWorkers(t *testing.T) {
	s := newScheduler()

	s.AddWorkers([]string{"localhost:9001", "localhost:9002", "localhost:9003"})

	errMsg := s.RemoveWorkers([]string{"localhost:9003"})
	if errMsg != "" {
		t.Fatal(errMsg)
	}

	totalWorkers := s.GetTotalWorkers()
	if totalWorkers != 2 {
		t.Errorf("Expected 2 workers but got %d instead", totalWorkers)
	}
}
