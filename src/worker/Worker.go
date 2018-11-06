package worker

import (
	"net/http"
	"net/url"
)

// WorkerConfig is a struct with parameters to construct new workers.
type WorkerConfig struct {
	URL    *url.URL
	Weight int
}

// Worker is an abstraction over worker nodes, it exposes methods that
// load-balancing algorithms can use to query information about the node's
// status and send workloads to it.
type Worker struct {
	url    *url.URL
	proxy  ReverseProxy
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

// SendWorkload sends a request to the node to run a workload. The worker's
// load increases by one unit until the work finishes.
func (worker *Worker) SendWorkload(w http.ResponseWriter, r *http.Request) {
	worker.load++
	worker.proxy.ProxyRequest(w, r)
	worker.load--
}

// NewWorker is a public constructor for Worker type. It receives a config
// struct and a ReverseProxy object to send workloads through the network.
func NewWorker(c WorkerConfig, p ReverseProxy) *Worker {
	return &Worker{url: c.URL, proxy: p, load: 0, weight: c.Weight}
}
