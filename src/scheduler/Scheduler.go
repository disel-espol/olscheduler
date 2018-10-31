package scheduler

import (
	"fmt"
	"net/http"

	"../balancer"
	"../httputil"
	"../lambda"
	"../schutil"
	"../worker"
)

// Scheduler is an object that can schedule lambda function workloads to a pool
// of workers.
type Scheduler struct {
	registry   map[string][]string
	myBalancer balancer.Balancer
	workers    []*worker.Worker
}

// NewScheduler is a public constructor for Scheduler
func NewScheduler(registry map[string][]string,
	myBalancer balancer.Balancer,
	workers []*worker.Worker) *Scheduler {
	return &Scheduler{registry, myBalancer, workers}
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
	for i, _ := range s.workers {
		s.workers[i].SendWorkload(w, r)
	}
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
	worker, err := s.myBalancer.SelectWorker(s.workers, r, lambda)
	if err != nil {
		httputil.RespondWithError(w, err)
	}

	worker.SendWorkload(w, r)
}

func (s *Scheduler) AddWorkers(urls []string) {
	for _, workerUrl := range urls {
		s.workers = schutil.AddWorkerToArray(s.workers, workerUrl, 1)
	}
}

func (s *Scheduler) GetTotalWorkers() int {
	return len(s.workers)
}

func (s *Scheduler) RemoveWorkers(urls []string) string {
	errMsg := ""
	for _, workerUrl := range urls {
		target := schutil.FindWorkerInArray(s.workers, "http://"+workerUrl)

		if target > -1 {
			s.workers = schutil.RemoveWorkerFromArray(s.workers, target)
		} else {
			errMsg += fmt.Sprintf("Unable to find worker with url: %s\n", workerUrl)
		}
	}
	return errMsg
}
