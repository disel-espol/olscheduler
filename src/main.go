package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"./balancer"
	"./schutil"
	"github.com/urfave/cli"
)

type Config struct {
	Host          string   `json:"host"`
	Port          int      `json:"port"`
	Balancer      string   `json:"balancer"`
	LoadThreshold int      `json:"load-threshold"`
	Workers       []string `json:"workers"`
}

// Global variables
var workers []schutil.Worker
var config Config
var supportedBalancers = []string{"random", "pkg-aware", "round-robin", "least-loaded"} // TODO(Gus): Change to a map[string]func

func DoRunLambda(w http.ResponseWriter, r *http.Request) (*schutil.Worker, *schutil.HttpError) {
	body, _ := ioutil.ReadAll(r.Body)
	strBody := string(body)
	pkgsRegExp := regexp.MustCompile(`"pkgs"\s*:\s*\[.*\],*\s*`)
	matches := pkgsRegExp.FindStringSubmatch(strBody)
	if len(matches) < 1 {
		return nil, &schutil.HttpError{"Pkgs array required", http.StatusInternalServerError}
	}
	strPkgsJson := matches[0]
	pkgsArrayRegExp := regexp.MustCompile(`\[.*\]`)
	srtPkgsMatches := pkgsArrayRegExp.FindStringSubmatch(strPkgsJson)
	if len(srtPkgsMatches) < 1 {
		return nil, &schutil.HttpError{"Pkgs array ill-formed", http.StatusInternalServerError}
	}
	decoder := json.NewDecoder(strings.NewReader(srtPkgsMatches[0]))
	var pkgs []string // pkgs ordered from larger to smaller
	err := decoder.Decode(&pkgs)
	if err != nil {
		return nil, &schutil.HttpError{err.Error(), http.StatusInternalServerError}
	}

	{ // Modify request's body
		newStrBody := strings.Replace(strBody, strPkgsJson, "", -1)
		r.Body = ioutil.NopCloser(strings.NewReader(newStrBody))
		r.ContentLength = int64(len(newStrBody))
	}

	{ // Select worker and serve http
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
			errorMsg := fmt.Sprintf("balancer: (%s) not supported. You could use one from %s", config.Balancer, supportedBalancers)
			return nil, &schutil.HttpError{errorMsg, http.StatusInternalServerError}
		}

		if err != nil {
			return nil, &schutil.HttpError{err.Error(), http.StatusInternalServerError}
		}
		for i, _ := range workers {
			log.Printf("--> worker: %s with load: %d", workers[i].URL.String(), workers[i].Load)
		}
		log.Printf("Selected Worker with URL: %s, balancer: %s, load-threshold: %d",
			worker.URL.String(),
			config.Balancer,
			config.LoadThreshold)
		worker.Load++
		worker.ReverseProxy.ServeHTTP(w, r)
		return worker, nil
	}
}

// RunLambda expects POST requests like this:
//
// curl -X POST localhost:9080/runLambda/<lambda-name> -d '{"pkgs": ["pkg0", "pkg1"], "param0": "value0"}'
func RunLambdaHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Receive request to %s\n", r.URL.Path)

	observer := schutil.NewObserverResponseWriter(w)
	worker, err := DoRunLambda(observer, r)
	if err != nil {
		log.Printf("Could not handle request: %s\n", err.Msg)
		http.Error(w, err.Msg, err.Code)
		return
	}
	worker.Load--
	log.Printf("Response Status: %d", observer.Status)
	// log.Printf("Response: %s", string(observer.Body))
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
	file, rfErr := ioutil.ReadFile(configFilepath)
	if rfErr != nil {
		log.Fatalf("Cannot read config file (%s)", configFilepath)
	}
	decoder := json.NewDecoder(bytes.NewReader(file))
	jsonErr := decoder.Decode(&config) // Parse json config file
	if jsonErr != nil {
		log.Fatalf("Config file Ill-formed (%s)", configFilepath)
	}

	if len(config.Workers)%2 != 0 {
		log.Fatalf("Config file Ill-formed (%s), every worker must have a weight", configFilepath)
	}

	for i := 0; i < len(config.Workers); i = i + 2 { // Make Workers with their Reverse Proxy Handlers
		weight, err := strconv.Atoi(config.Workers[i+1])
		if err != nil || weight < 0 {
			log.Fatalf("Config file Ill-formed (%s), every worker weight must be a positive number", configFilepath)
		}
		u, _ := url.Parse("http://" + config.Workers[i])
		handler := httputil.NewSingleHostReverseProxy(u)
		workers = append(workers, schutil.Worker{u, handler, 0, weight})
	}

	// http.HandleFunc("/scheduler", func(w http.ResponseWriter, r *http.Request) {
	//     fmt.Fprint(w, "I am a humble scheduler")
	// })
	http.HandleFunc("/runLambda/", RunLambdaHandler)
	http.HandleFunc("/status", StatusHandler)
	rand.Seed(time.Now().Unix()) // For rand future calls
	log.Print("Scheduler is running")
	http.ListenAndServe(fmt.Sprintf("%s:%d", config.Host, config.Port), nil)
}

func main() {
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
	app.Run(os.Args)
}
