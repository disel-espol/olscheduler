package main

import (
	"os"

	"./client"
	"./schutil"
	"./server"

	"github.com/urfave/cli"
)

func createCliApp() *cli.App {
	app := cli.NewApp()
	app.Usage = "Scheduler for Open-Lambda"
	app.UsageText = "olscheduler COMMAND [ARG...]"
	app.ArgsUsage = "ArgsUsage"
	app.EnableBashCompletion = true
	app.HideVersion = true

	configFlag := cli.StringFlag{
		Name:  "config, c",
		Usage: "Config json file",
		Value: "olscheduler.json",
	}
	app.Commands = []cli.Command{
		cli.Command{Name: "start", Usage: "Start Open-Lambda Scheduler",
			UsageText:   "olscheduler start [-c|--config=FILEPATH]",
			Description: "The scheduler starts with settings from config json file.",
			Flags:       []cli.Flag{configFlag},
			Action: func(c *cli.Context) error {
				configFilepath := c.String("config")
				config := schutil.LoadConfigFromFile(configFilepath)
				return server.Start(configFilepath, config)
			},
		},
		cli.Command{
			Name:  "workers",
			Usage: "Worker nodes management",
			Subcommands: []cli.Command{
				{
					Name:      "add",
					Usage:     "add a new worker node to an already running scheduler",
					UsageText: "olscheduler worker add URL",
					Flags:     []cli.Flag{configFlag},
					Action: func(c *cli.Context) error {
						configFilepath := c.String("config")
						config := schutil.LoadConfigFromFile(configFilepath)
						return client.AddWorkers(config, c.Args())
					},
				},
				{
					Name:      "remove",
					Usage:     "remove an existing worker node from an already running scheduler",
					UsageText: "olscheduler worker remove URL",
					Flags:     []cli.Flag{configFlag},
					Action: func(c *cli.Context) error {
						configFilepath := c.String("config")
						config := schutil.LoadConfigFromFile(configFilepath)
						return client.RemoveWorkers(config, c.Args())
					},
				},
			},
		},
	}
	return app
}

func main() {
	app := createCliApp()
	app.Run(os.Args)
}
