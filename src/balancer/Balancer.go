package balancer

import "net/http"

type Balancer interface {
	SelectWorker(workers []schutil.Worker, r http.Request) (*schutil.Worker, error)
}
