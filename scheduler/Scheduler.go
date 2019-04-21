package scheduler

import (
	"fmt"
	"net/http"

	"github.com/disel-espol/olscheduler/balancer"
	"github.com/disel-espol/olscheduler/config"
	"github.com/disel-espol/olscheduler/httputil"
	"github.com/disel-espol/olscheduler/lambda"
	"github.com/disel-espol/olscheduler/worker"
)

// Scheduler is an object that can schedule lambda function workloads to a pool
// of workers.
type Scheduler struct {
	registry   map[string][]string
	myBalancer balancer.Balancer
	proxy      worker.ReverseProxy
	pool       worker.WorkerPool
}

func (s *Scheduler) GetLambdaInfoFromRequest(r *http.Request) (*lambda.Lambda,
	*httputil.HttpError) {
	lambdaName := httputil.Get2ndPathSegment(r, "runLambda")

	if lambdaName == "" {
		return nil, &httputil.HttpError{
			fmt.Sprintf("Could not find lambda name in path %s", r.URL.Path),
			http.StatusBadRequest}
	}

	pkgs, found := s.registry[lambdaName]
	if !found {
		return nil, &httputil.HttpError{
			fmt.Sprintf("No pkgs found in registry for lambda name: %s",
				lambdaName),
			http.StatusBadRequest}
	}

	return &lambda.Lambda{lambdaName, pkgs}, nil
}

func (s *Scheduler) SendToAllWorkers(w http.ResponseWriter, r *http.Request) {
	workers := s.pool.GetWorkers()
	for i, _ := range workers {
		go s.sendWorkload(workers[i], w, r)
	}
}

func (s *Scheduler) sendWorkload(
	selectedWorker *worker.Worker,
	w http.ResponseWriter,
	r *http.Request) {

	s.pool.UpdateWorkerLoad(selectedWorker, true)
	s.proxy.ProxyRequest(selectedWorker, w, r)
	s.pool.UpdateWorkerLoad(selectedWorker, false)
}

// RunLambda is an HTTP request handler that expects requests of form
// /runLambda/<lambdaName>. It extracts the lambda name from the request path
// and then chooses a worker to run the lambda workload using the configured
// load balancer. The lambda response is forwarded to the client "as-is"
// without any modifications.
func (s *Scheduler) RunLambda(w http.ResponseWriter, r *http.Request) {
	lambda, err := s.GetLambdaInfoFromRequest(r)

	if err != nil {
		httputil.RespondWithError(w, err)
		return
	}

	// Select worker and serve http
	workers := s.pool.GetWorkers()
	selectedWorker, err := s.myBalancer.SelectWorker(workers, r, lambda)
	if err != nil {
		httputil.RespondWithError(w, err)
	}
	s.sendWorkload(selectedWorker, w, r)

	if c, ok := s.myBalancer.(balancer.BalancerReleaser); ok {
		c.ReleaseWorker(selectedWorker.GetURL())
	}
}

func (s *Scheduler) AddWorkers(urls []string) {
	for _, workerUrl := range urls {
		s.pool.AddWorker(workerUrl, 1)
	}
}

func (s *Scheduler) GetTotalWorkers() int {
	return s.pool.GetTotalWorkers()
}

func (s *Scheduler) RemoveWorkers(urls []string) string {
	errMsg := ""
	for _, workerUrl := range urls {
		target := s.pool.FindWorker("http://" + workerUrl)

		if target > -1 {
			s.pool.RemoveWorkerAt(target)
		} else {
			errMsg += fmt.Sprintf("Unable to find worker with url: %s\n", workerUrl)
		}
	}
	return errMsg
}

func (s *Scheduler) ManagePool() {
	s.pool.ManagePool()
}

func NewScheduler(c config.Config) *Scheduler {
	return &Scheduler{
		c.Registry,
		c.Balancer,
		c.ReverseProxy,
		worker.NewWorkerPool(c.Workers),
	}

}
