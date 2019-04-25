package server

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/disel-espol/olscheduler/config"
	"github.com/disel-espol/olscheduler/httputil"
	"github.com/disel-espol/olscheduler/scheduler"
)

var myScheduler *scheduler.Scheduler
var myConfig config.Config

func parseWorkerURLs(querySlice []string) ([]url.URL, *httputil.HttpError) {
	totalWorkers := len(querySlice)
	if totalWorkers < 1 {
		return nil, httputil.New400Error("Workers array in query string cannot be empty")
	}

	workerUrls := make([]url.URL, totalWorkers)
	for i, urlString := range querySlice {
		workerUrl, parseErr := url.Parse(urlString)
		if parseErr != nil {
			return nil, httputil.New400Error("Malformed worker URL: " + urlString)
		}
		workerUrls[i] = *workerUrl
	}
	return workerUrls, nil
}

// RunLambda expects POST requests like this:
//
// curl -X POST localhost:9080/runLambda/<lambda-name> -d '{"param0": "value0"}'
func runLambdaHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Receive request to %s\n", r.URL.Path)

	observer := httputil.NewObserverResponseWriter(w)
	myScheduler.RunLambda(observer, r)

	log.Printf("Response Status: %d", observer.Status)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	appendResponseWriter := httputil.NewAppendResponseWriter()
	myScheduler.StatusCheckAllWorkers(appendResponseWriter, r)
}

func addWorkerHandler(w http.ResponseWriter, r *http.Request) {
	workers := r.URL.Query()["workers"]

	workerUrls, err := parseWorkerURLs(workers)
	if err != nil {
		httputil.RespondWithError(w, err)
		return
	}
	myScheduler.AddWorkers(workerUrls)
}

func removeWorkerHandler(w http.ResponseWriter, r *http.Request) {
	workers := r.URL.Query()["workers"]

	workerUrls, err := parseWorkerURLs(workers)
	if err != nil {
		httputil.RespondWithError(w, err)
		return
	}
	myScheduler.RemoveWorkers(workerUrls)
}

func Start(c config.Config) error {
	myConfig = c
	myScheduler = scheduler.NewScheduler(c)

	http.HandleFunc("/runLambda/", runLambdaHandler)
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/admin/workers/add", addWorkerHandler)
	http.HandleFunc("/admin/workers/remove", removeWorkerHandler)

	url := fmt.Sprintf("%s:%d", myConfig.Host, myConfig.Port)
	return http.ListenAndServe(url, nil)
}
