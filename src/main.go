package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"./balancer"
	"./httputil"
	"./scheduler"
	"./schutil"

	"github.com/urfave/cli"
)

// Global variables
var myScheduler *scheduler.Scheduler
var config schutil.Config

// RunLambda expects POST requests like this:
//
// curl -X POST localhost:9080/runLambda/<lambda-name> -d '{"param0": "value0"}'
func RunLambdaHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Receive request to %s\n", r.URL.Path)

	observer := httputil.NewObserverResponseWriter(w)
	myScheduler.RunLambda(observer, r)

	log.Printf("Response Status: %d", observer.Status)
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	appendResponseWriter := httputil.NewAppendResponseWriter()
	myScheduler.SendToAllWorkers(appendResponseWriter, r)
	fmt.Fprint(w, string(appendResponseWriter.Body))
}

func StartServer(ctx *cli.Context) {
	configFilepath := ctx.String("config")
	config = schutil.LoadConfigFromFile(configFilepath)
	myBalancer := balancer.CreateBalancerFromConfig(config)
	workers := schutil.CreateWorkersArray(configFilepath, config)
	registry := schutil.CreateRegistryFromFile(config.Registry)
	myScheduler = scheduler.NewScheduler(registry, myBalancer, workers)

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
