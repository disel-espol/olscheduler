package balancer

import (
	"../schutil"
	"errors"
)

var nextIndex = 0

func SelectWorkerRoundRobin(workers []schutil.Worker) (*schutil.Worker, error) {
	if len(workers) == 0 {
		return nil, errors.New("Can't select worker, Workers empty")
	}

	currentIndex := nextIndex
	
	nextIndex++
	if (nextIndex >= len(workers)) {
		nextIndex = 0
	}

	return &workers[currentIndex], nil
}