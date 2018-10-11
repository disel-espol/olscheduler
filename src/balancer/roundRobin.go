package balancer

import (
	"../schutil"
	"errors"
	"http"
)

type RoundRobinBalancer struct {
	nextIndex int
}

func (b *RoundRobinBalancer) SelectWorker(workers []schutil.Worker, r http.Request) {
	if len(workers) == 0 {
		return nil, errors.New("Can't select worker, Workers empty")
	}

	currentIndex := b.nextIndex

	b.nextIndex++
	if b.nextIndex >= len(workers) {
		nextIndex = 0
	}

	return &workers[currentIndex], nil
}
