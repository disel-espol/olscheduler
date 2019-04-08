package balancer

import (
	"net/http"

	"../httputil"
	"../lambda"
	"../worker"
)

type Balancer interface {
	SelectWorker(workers []*worker.Worker, r *http.Request, l *lambda.Lambda) (*worker.Worker, *httputil.HttpError)
}

type BalancerReleaser interface {
	ReleaseWorker(host string)
}

