package balancer

import (
	"net/http"

	"../schutil"
	"../worker"
)

type Balancer interface {
	SelectWorker(workers []*worker.Worker, r *http.Request) (*worker.Worker, *schutil.HttpError)
}
