package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"./balancer"
	"./httpreq"
	"./schutil"
	"github.com/urfave/cli"
)

// Global variables
var workers []schutil.Worker
var config schutil.Config
var registry map[string][]string
var balancer balancer.Balancer

func logSelectedWorker(worker *schutil.Worker) {
	for i, _ := range workers {
		log.Printf("--> worker: %s with load: %d", workers[i].URL.String(), workers[i].GetLoad())
	}
	log.Printf("Selected Worker with URL: %s, balancer: %s, load-threshold: %d",
		worker.URL.String(),
		config.Balancer,
		config.LoadThreshold)
}

// Copied from OpenLamda src
// getUrlComponents parses request URL into its "/" delimated components
func getUrlComponents(r *http.Request) []string {
	path := r.URL.Path

	// trim prefix
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}

	// trim trailing "/"
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	components := strings.Split(path, "/")
	return components
}

func DoRunLambdaPlus(w http.ResponseWriter, r *http.Request) *schutil.HttpError {
	{ // Change request's url
		newPath := r.URL.Path
		newPath = strings.Replace(newPath, "runLambdaPlus", "runLambda", 1)
		r.URL.Path = newPath
	}
	strBody := httpreq.GetBodyAsString(r)
	pkgs, newStrBody, err := stripPkgsArrayFromBody(strBody)

	if err != nil {
		return err
	}

	httpreq.ReplaceBodyWithString(r, newStrBody)

	{ // Select worker and serve http
		var worker *schutil.Worker
		worker, err = selectWorkerUsingConfiguredBalancer(pkgs)
		if err != nil {
			return err
		}

		logSelectedWorker(worker)
		worker.SendWorkload(w, r)
		return nil
	}
}

// RunLambda expects POST requests like this:
//
// curl -X POST localhost:9080/runLambdaPlus/<lambda-name> -d '{"pkgs": ["pkg0", "pkg1"], "param0": "value0"}'
func RunLambdaPlusHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Receive request to %s\n", r.URL.Path)

	observer := schutil.NewObserverResponseWriter(w)
	err := DoRunLambdaPlus(observer, r)
	if err != nil {
		log.Printf("Could not handle request: %s\n", err.Msg)
		http.Error(w, err.Msg, err.Code)
		return
	}
	log.Printf("Response Status: %d", observer.Status)
}

func DoRunLambda(w http.ResponseWriter, r *http.Request) {
	var pkgs []string
	var found bool
	{ // Try to load from registry
		lambdaName := ""
		{ // Copied from OpenLamda src
			urlParts := getUrlComponents(r)
			if len(urlParts) < 2 {
				return &schutil.HttpError{"Name of image to run required", http.StatusBadRequest}
			}
			img := urlParts[1]
			i := strings.Index(img, "?")
			if i >= 0 {
				img = img[:i-1]
			}
			lambdaName = img
		}
		// fmt.Println(lambdaName)
		pkgs, found = registry[lambdaName]
		if !found {
			return &schutil.HttpError{fmt.Sprintf("No pkgs found in registry: %v for lambda name: %v", config.Registry, lambdaName), http.StatusBadRequest}
		}
		// fmt.Println(pkgs)
	}

	{ // Select worker and serve http
		worker, err := myBalancer.SelectWorker(workers, r)
		if err != nil {
			respondWithError(w, err)
		}

		logSelectedWorker(worker)
		worker.SendWorkload(w, r)
	}
}

// RunLambda expects POST requests like this:
//
// curl -X POST localhost:9080/runLambda/<lambda-name> -d '{"param0": "value0"}'
func RunLambdaHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Receive request to %s\n", r.URL.Path)

	observer := schutil.NewObserverResponseWriter(w)
	err := DoRunLambda(observer, r)
	if err != nil {
		log.Printf("Could not handle request: %s\n", err.Msg)
		http.Error(w, err.Msg, err.Code)
		return
	}
	log.Printf("Response Status: %d", observer.Status)
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	appendResponseWriter := schutil.NewAppendResponseWriter()
	for i, _ := range workers {
		workers[i].ReverseProxy.ServeHTTP(appendResponseWriter, r)
	}
	fmt.Fprint(w, string(appendResponseWriter.Body))
}

func StartServer(ctx *cli.Context) {
	configFilepath := ctx.String("config")
	config = schutil.LoadConfigFromFile(configFilepath)
	balancer = balancer.CreateBalancerFromConfig(config)
	workers = schutil.CreateWorkersArray(configFilepath, config)
	registry = schutil.CreateRegistryFromFile(config.Registry)

	http.HandleFunc("/runLambdaPlus/", RunLambdaPlusHandler)
	http.HandleFunc("/runLambda/", RunLambdaHandler)
	http.HandleFunc("/status", StatusHandler)
	log.Print("Scheduler is running")
	http.ListenAndServe(fmt.Sprintf("%s:%d", config.Host, config.Port), nil)
}

func createCliApp() *cli.App {
	app := cli.NewApp()
	app.Usage = "Scheduler for Open-Lambda"
	app.UsageText = "olscheduler COMMAND [ARG...]"
	app.ArgsUsage = "ArgsUsage"
	app.EnableBashCompletion = true
	app.HideVersion = true
	app.Commands = []cli.Command{
		cli.Command{
			Name:        "start",
			Usage:       "Start Open-Lambda Scheduler",
			UsageText:   "olscheduler start [-c|--config=FILEPATH]",
			Description: "The scheduler starts with settings from config json file.",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config, c",
					Usage: "Config json file",
					Value: "olscheduler.json",
				},
			},
			Action: StartServer,
		},
	}
	return app
}

func main() {
	app := createCliApp()
	app.Run(os.Args)
}
