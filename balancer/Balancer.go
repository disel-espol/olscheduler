package balancer

import (
	"net/http"

	"github.com/disel-espol/olscheduler/httputil"
	"github.com/disel-espol/olscheduler/lambda"
	"github.com/disel-espol/olscheduler/worker"
)

type Balancer interface {
	SelectWorker(workers []*worker.Worker, r *http.Request, l *lambda.Lambda) (*worker.Worker, *httputil.HttpError)
}

type BalancerReleaser interface {
	ReleaseWorker(host string)
}
