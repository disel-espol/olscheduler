package worker

import (
	"net/url"
)

// Worker is a model for worker nodes, it exposes methods that
// load-balancing algorithms can use to query information about the node's
// status
type Worker struct {
	url    *url.URL
	load   int
	weight int
}

func (worker *Worker) GetURL() string {
	return worker.url.String()
}

func (worker *Worker) GetLoad() int {
	return worker.load
}

func (worker *Worker) GetWeight() int {
	return worker.weight
}

// NewWorker is a public constructor for Worker type.
func NewWorker(url *url.URL, weight int) *Worker {
	return &Worker{url: url, load: 0, weight: weight}
}
