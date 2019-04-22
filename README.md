# olscheduler [![CircleCI](https://circleci.com/gh/disel-espol/olscheduler.svg?style=svg)](https://circleci.com/gh/disel-espol/olscheduler)

Extensible scheduler for OpenLambda written in Go.

## Installation

Once you have the Go runtime installed in your system just run:

``` bash
go get github.com/disel-espol/olscheduler
```

## Usage 

This program launches an HTTP server that proxies incoming HTTP requests to the
HTTP servers that will actually handle the requests, known as "worker nodes". 
Please note that it is your job to launch the worker nodes and manage them. You 
can configure almost everything about the scheduler via a Go API or a JSON 
configuration file using the CLI API.

### CLI API

First you must create a configuration file. Here's an example of a configuration
to run a scheduler listening on port 9020 using the "Package Aware" load 
balancing algorithm with two worker nodes listening on ports 9021, and 9022 
respectively.

``` JSON
{
  "host":"localhost",
  "port":9020,
  "load-threshold":3,
  "registry":"/tmp/olscheduler-registry.json",
  "balancer":"pkg-aware",
  "workers":[
    "localhost:9021", "1",
    "localhost:9022" ,"1"
  ]
}
```

`/tmp/olscheduler-registry.json` is another JSON file with an array information 
about the handler functions. This data is currently only used by the 
"Package Aware" algorithm. Each item in the array has two attributes:

- `handle`. The name of the handler funcion.
- `pkgs`. An array of unique identifier names for the packages that the handler 
  function requires as dependencies.

Here's an example of a registry file with two handler functions: `foo` and `bar`.

``` JSON
[
  {
    "handle":"foo",
    "pkgs":["pkg0","pkg1"]
  },
  {
    "handle":"bar",
    "pkgs":["pkg0","pkg2"]
  }
]
```

Once the files are ready you can launch the scheduler by running:

``` bash
olscheduler start -c /path/to/config/file 
```

### Go API

This is the simplest method, provided that you are comfortable with Go. Just 
create a `Config` struct with the desired configuration and then pass it to the 
`server.Start()` function to start the server.

``` Go
package main

import (
	"github.com/disel-espol/olscheduler/config"
	"github.com/disel-espol/olscheduler/server"
	"github.com/disel-espol/olscheduler/worker"
	"net/url"
)

func main() {
	myConfig := config.CreateDefaultConfig()

	// edit config as you wish
	myConfig.Port = 8080
	myConfig.LoadThreshold = 5

  // add functions to registry
	myConfig.Registry["foo"] = []string{"pkg0", "pkg1"}
	myConfig.Registry["bar"] = []string{"pkg0", "pkg2"}

  // add worker nodes
	workerUrl1, _ := url.Parse("http://localhost:8081")
	workerUrl2, _ := url.Parse("http://localhost:8082")
	myConfig.Workers = []*worker.Worker{
		worker.NewWorker(workerUrl1, 1),
		worker.NewWorker(workerUrl2, 1),
	}

  // launch the server
	server.Start(myConfig)
}
```


