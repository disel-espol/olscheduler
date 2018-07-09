package balancer

import (
	"../schutil"
	"errors"
)

func SelectWorkerLeastLoaded(workers []schutil.Worker) (*schutil.Worker, error) {
	if len(workers) == 0 {
		return nil, errors.New("Can't select worker, Workers empty")
	}

	targetIndex := 0
	for i := 1; i < len(workers); i++ {
		if workers[i].GetLoad() < workers[targetIndex].GetLoad() {
			targetIndex = i
		}
	}

	return &workers[targetIndex], nil
}
