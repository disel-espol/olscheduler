package server

import (
	"fmt"
	"log"
	"net/http"

	"../balancer"
	"../httputil"
	"../scheduler"
	"../schutil"
)

var myScheduler *scheduler.Scheduler
var config schutil.Config

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
	myScheduler.SendToAllWorkers(appendResponseWriter, r)
	fmt.Fprint(w, string(appendResponseWriter.Body))
}

func addWorkerHandler(w http.ResponseWriter, r *http.Request) {
	workers := r.URL.Query()["workers"]

	log.Printf("Got request to remove workers %v full %v", workers, r.URL.Query())
	if len(workers) < 1 {
		err := httputil.New400Error("Workers array in query string cannot be empty")
		httputil.RespondWithError(w, err)
		return
	}

	myScheduler.AddWorkers(workers)
}

func removeWorkerHandler(w http.ResponseWriter, r *http.Request) {
	workers := r.URL.Query()["workers"]

	if len(workers) < 1 {
		err := httputil.New400Error("Workers array in query string cannot be empty")
		httputil.RespondWithError(w, err)
		return
	}

	errMsg := myScheduler.RemoveWorkers(workers)
	if errMsg != "" {
		err := httputil.New400Error(errMsg)
		httputil.RespondWithError(w, err)
		return
	}
}

func Start(configFilepath string, config schutil.Config) error {
	myBalancer := balancer.CreateBalancerFromConfig(config)
	workers := schutil.CreateWorkersArray(configFilepath, config)
	registry := schutil.CreateRegistryFromFile(config.Registry)
	myScheduler = scheduler.NewScheduler(registry, myBalancer, workers)

	http.HandleFunc("/runLambda/", runLambdaHandler)
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/admin/workers/add", addWorkerHandler)
	http.HandleFunc("/admin/workers/remove", removeWorkerHandler)
	log.Print("Scheduler is running")
	return http.ListenAndServe(fmt.Sprintf("%s:%d", config.Host, config.Port), nil)
}
