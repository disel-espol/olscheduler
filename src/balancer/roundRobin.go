package balancer

import (
	"net/http"

	"../schutil"
	"../worker"
)

type RoundRobinBalancer struct {
	nextIndex int
}

func (b *RoundRobinBalancer) SelectWorker(workers []*worker.Worker, r *http.Request) (*worker.Worker, *schutil.HttpError) {
	if len(workers) == 0 {
		return nil, schutil.New500Error("Can't select worker, Workers empty")
	}

	currentIndex := b.nextIndex

	b.nextIndex++
	if b.nextIndex >= len(workers) {
		b.nextIndex = 0
	}

	return workers[currentIndex], nil
}
