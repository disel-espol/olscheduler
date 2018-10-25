package balancer

import (
	"net/http"

	"../httputil"
	"../lambda"
	"../worker"
)

type RoundRobinBalancer struct {
	nextIndex int
}

func (b *RoundRobinBalancer) SelectWorker(workers []*worker.Worker, r *http.Request, l *lambda.Lambda) (*worker.Worker, *httputil.HttpError) {
	if len(workers) == 0 {
		return nil, httputil.New500Error("Can't select worker, Workers empty")
	}

	currentIndex := b.nextIndex

	b.nextIndex++
	if b.nextIndex >= len(workers) {
		b.nextIndex = 0
	}

	return workers[currentIndex], nil
}
