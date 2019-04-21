package balancer

import (
	"math/rand"
	"net/http"

	"github.com/disel-espol/olscheduler/httputil"
	"github.com/disel-espol/olscheduler/lambda"
	"github.com/disel-espol/olscheduler/worker"
)

// Select a random worker
// assumes that rand.Seed(time.Now().Unix()) has already been called
type RandomBalancer struct {
}

func (b *RandomBalancer) SelectWorker(workers []*worker.Worker, r *http.Request, l *lambda.Lambda) (*worker.Worker, *httputil.HttpError) {
	if len(workers) == 0 {
		return nil, httputil.New500Error("Can't select worker, Workers empty")
	}
	totalWeight := 0
	for i, _ := range workers {
		if workers[i].GetWeight() < 0 {
			return nil, httputil.New500Error("Worker's Weight cannot be negative")
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
		return nil, httputil.New500Error("Can't select worker, All weights are zero")
	}

	return workers[targetIndex], nil
}
