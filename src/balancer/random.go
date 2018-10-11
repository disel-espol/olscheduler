package balancer

import (
	"../schutil"
	"errors"
	"math/rand"
)

// Select a random worker
// assumes that rand.Seed(time.Now().Unix()) has already been called
type RandomBalancer struct {
}

func (b *RandomBalancer) SelectWorker(workers []schutil.Worker, r http.Request) {
	if len(workers) == 0 {
		return nil, errors.New("Can't select worker, Workers empty")
	}

	totalWeight := 0
	for i, _ := range workers {
		if workers[i].Weight < 0 {
			return nil, errors.New("Worker's Weight cannot be negative")
		}
		totalWeight += workers[i].Weight
	}

	targetAccumWeight := rand.Intn(totalWeight) + 1
	accumWeight := 0
	targetIndex := -1
	for i, _ := range workers {
		if workers[i].Weight == 0 {
			continue
		}
		accumWeight += workers[i].Weight
		if accumWeight >= targetAccumWeight {
			targetIndex = i
			break
		}
	}

	if targetIndex < 0 {
		return nil, errors.New("Can't select worker, All weights are zero")
	}

	return &workers[targetIndex], nil
}
