package scheduler

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/disel-espol/olscheduler/balancer"
	"github.com/disel-espol/olscheduler/config"
	"github.com/disel-espol/olscheduler/httputil"
	"github.com/disel-espol/olscheduler/lambda"
	"github.com/disel-espol/olscheduler/proxy"
)

// Scheduler is an object that can schedule lambda function workloads to a pool
// of workers.
type Scheduler struct {
	registry map[string][]string
	balancer balancer.Balancer
	proxy    proxy.ReverseProxy
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

func (s *Scheduler) StatusCheckAllWorkers(w http.ResponseWriter, r *http.Request) {
	for _, workerUrl := range s.balancer.GetAllWorkers() {
		s.proxy.ProxyRequest(workerUrl, w, r)
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
	selectedWorkerURL, err := s.balancer.SelectWorker(r, lambda)
	if err != nil {
		httputil.RespondWithError(w, err)
		return
	}

	s.proxy.ProxyRequest(selectedWorkerURL, w, r)
	s.balancer.ReleaseWorker(selectedWorkerURL)
}

func (s *Scheduler) AddWorkers(urls []url.URL) {
	for _, workerURL := range urls {
		s.balancer.AddWorker(workerURL)
	}
}
func (s *Scheduler) RemoveWorkers(urls []url.URL) {
	for _, workerURL := range urls {
		s.balancer.RemoveWorker(workerURL)
	}
}

func NewScheduler(c config.Config) *Scheduler {
	return &Scheduler{
		c.Registry,
		c.Balancer,
		c.ReverseProxy,
	}

}
