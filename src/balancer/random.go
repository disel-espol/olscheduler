package balancer

import (
	"math/rand"
	"net/http"

	"../schutil"
	"../worker"
)

// Select a random worker
// assumes that rand.Seed(time.Now().Unix()) has already been called
type RandomBalancer struct {
}

func (b *RandomBalancer) SelectWorker(workers []*worker.Worker, r *http.Request) (*worker.Worker, *schutil.HttpError) {
	if len(workers) == 0 {
		return nil, schutil.New500Error("Can't select worker, Workers empty")
	}
	totalWeight := 0
	for i, _ := range workers {
		if workers[i].GetWeight() < 0 {
			return nil, schutil.New500Error("Worker's Weight cannot be negative")
		}
		totalWeight += workers[i].GetWeight()
	}

	targetAccumWeight := rand.Intn(totalWeight) + 1
	accumWeight := 0
	targetIndex := -1
	for i, _ := range workers {
		if workers[i].GetWeight() == 0 {
			continue
		}
		accumWeight += workers[i].GetWeight()
		if accumWeight >= targetAccumWeight {
			targetIndex = i
			break
		}
	}

	if targetIndex < 0 {
		return nil, schutil.New500Error("Can't select worker, All weights are zero")
	}

	return workers[targetIndex], nil
}
