package balancer

import (
	"net/http"

	"github.com/disel-espol/olscheduler/httputil"
	"github.com/disel-espol/olscheduler/lambda"
	"github.com/disel-espol/olscheduler/worker"
)

type LeastLoadedBalancer struct {
}

func (b *LeastLoadedBalancer) SelectWorker(workers []*worker.Worker, r *http.Request, l *lambda.Lambda) (*worker.Worker, *httputil.HttpError) {
	if len(workers) == 0 {
		return nil, httputil.New500Error("Can't select worker, Workers empty")
	}

	targetIndex := 0
	for i := 1; i < len(workers); i++ {
		if workers[i].GetLoad() < workers[targetIndex].GetLoad() {
			targetIndex = i
		}
	}

	return workers[targetIndex], nil
}
