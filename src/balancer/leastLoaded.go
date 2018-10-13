package balancer

import (
	"net/http"

	"../schutil"
	"../worker"
)

type LeastLoadedBalancer struct {
}

func (b *LeastLoadedBalancer) SelectWorker(workers []*worker.Worker, r *http.Request) (*worker.Worker, *schutil.HttpError) {
	if len(workers) == 0 {
		return nil, schutil.New500Error("Can't select worker, Workers empty")
	}

	targetIndex := 0
	for i := 1; i < len(workers); i++ {
		if workers[i].GetLoad() < workers[targetIndex].GetLoad() {
			targetIndex = i
		}
	}

	return workers[targetIndex], nil
}
