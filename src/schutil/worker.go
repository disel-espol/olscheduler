package schutil

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Worker struct {
	URL          *url.URL
	ReverseProxy *httputil.ReverseProxy
	load         int
	Weight       int
}

func (worker *Worker) GetLoad() int {
	return worker.load
}

func (worker *Worker) SendWorkload(w http.ResponseWriter, r *http.Request) {
	worker.load++
	worker.ReverseProxy.ServeHTTP(w, r)
	worker.load--
}
