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
var supportedBalancers = []string{"random", "pkg-aware", "round-robin", "least-loaded"} // TODO(Gus): Change to a map[string]func

func stripPkgsArrayFromBody(strBody string) ([]string, string, *schutil.HttpError) {
	pkgsRegExp := regexp.MustCompile(`"pkgs"\s*:\s*\[.*\],*\s*`)
	matches := pkgsRegExp.FindStringSubmatch(strBody)
	if len(matches) < 1 {
		return nil, strBody, nil
	}
	strPkgsJson := matches[0]
	pkgsArrayRegExp := regexp.MustCompile(`\[.*\]`)
	srtPkgsMatches := pkgsArrayRegExp.FindStringSubmatch(strPkgsJson)
	if len(srtPkgsMatches) < 1 {
		return nil, "", &schutil.HttpError{"Pkgs array ill-formed", http.StatusInternalServerError}
	}
	decoder := json.NewDecoder(strings.NewReader(srtPkgsMatches[0]))
	var pkgs []string // pkgs ordered from larger to smaller
	err := decoder.Decode(&pkgs)
	if err != nil {
		return nil, "", &schutil.HttpError{err.Error(), http.StatusInternalServerError}
	}

	newStrBody := strings.Replace(strBody, strPkgsJson, "", -1)

	return pkgs, newStrBody, nil
}

// If we could make a "Balancer" interface and have all algorithms receive the
// same parameter list we could completely delete this function
func selectWorkerUsingConfiguredBalancer(pkgs []string) (*schutil.Worker, *schutil.HttpError) {
	var worker *schutil.Worker
	var err error
	switch config.Balancer {
	case supportedBalancers[0]:
		worker, err = balancer.SelectWorkerRandom(workers)
	case supportedBalancers[1]:
		worker, err = balancer.SelectWorkerPkgAware(workers, pkgs, config.LoadThreshold)
	case supportedBalancers[2]:
		worker, err = balancer.SelectWorkerRoundRobin(workers)
	case supportedBalancers[3]:
		worker, err = balancer.SelectWorkerLeastLoaded(workers)
	default:
		errorMsg := fmt.Sprintf("balancer: (%s) not supported. You could use one from %s",
			config.Balancer, supportedBalancers)
		return nil, &schutil.HttpError{errorMsg, http.StatusInternalServerError}
	}

	if err != nil {
		return nil, &schutil.HttpError{err.Error(), http.StatusInternalServerError}
	}

	return worker, nil
}

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

func DoRunLambda(w http.ResponseWriter, r *http.Request) *schutil.HttpError {
	strBody := httpreq.GetBodyAsString(r)
	pkgs, newStrBody, err := stripPkgsArrayFromBody(strBody)

	if err != nil {
		return err
	}

	httpreq.ReplaceBodyWithString(r, newStrBody)

	if pkgs == nil { // Try to load from registry
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
		pkgs = registry[lambdaName]
		// fmt.Println(pkgs)
	}

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
// curl -X POST localhost:9080/runLambda/<lambda-name> -d '{"pkgs": ["pkg0", "pkg1"], "param0": "value0"}'
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
	workers = schutil.CreateWorkersArray(configFilepath, config)
	registry = schutil.CreateRegistryFromFile(config.Registry)

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
	rand.Seed(time.Now().Unix()) // For rand future calls

	app := createCliApp()
	app.Run(os.Args)
}
