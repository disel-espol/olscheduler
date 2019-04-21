package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/url"
	"strconv"

	"github.com/disel-espol/olscheduler/worker"
)

// JSONConfig holds the data configured via a JSON file. This shall be used
// to parse the JSON file and create a proper Config struct that dictates the
// scheduler's behavior.
type JSONConfig struct {
	Host          string   `json:"host"`
	Port          int      `json:"port"`
	Balancer      string   `json:"balancer"`
	LoadThreshold int      `json:"load-threshold"`
	Registry      string   `json:"registry"`
	Workers       []string `json:"workers"`
}

type Handle struct {
	Handle string   `json:handle`
	Pkgs   []string `json:pkgs`
}

func (c JSONConfig) ToConfig() Config {
	return Config{
		Host:          c.Host,
		Port:          c.Port,
		LoadThreshold: c.LoadThreshold,
		Balancer:      createBalancerFromConfig(c),
		Workers:       createWorkerSlice(c),
		Registry:      createRegistryFromFile(c.Registry),
		ReverseProxy:  worker.NewHTTPReverseProxy(),
	}
}

func LoadConfigFromFile(configFilepath string) JSONConfig {
	var config JSONConfig
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

	return config
}

func createWorkerSlice(config JSONConfig) []*worker.Worker {
	workersLength := len(config.Workers)
	if workersLength%2 == 1 {
		log.Fatalf("Config file Ill-formed, every worker url must be followed by its weight")
	}

	workerSlice := make([]*worker.Worker, workersLength/2)
	for i := 0; i < workersLength; i = i + 2 {
		weight, err := strconv.Atoi(config.Workers[i+1])
		if err != nil || weight < 0 {
			log.Fatalf("Config file Ill-formed, every worker weight must be a positive number")
		}
		workerUrl, _ := url.Parse("http://" + config.Workers[i])
		workerSlice[i/2] = worker.NewWorker(workerUrl, weight)
	}

	return workerSlice
}

func createRegistryFromFile(registryFilePath string) map[string][]string {
	var handles []Handle
	file, rfErr := ioutil.ReadFile(registryFilePath)
	if rfErr != nil {
		log.Fatalf("Cannot read registry file (%s)", registryFilePath)
	}
	decoder := json.NewDecoder(bytes.NewReader(file))
	jsonErr := decoder.Decode(&handles) // Parse json registry file
	if jsonErr != nil {
		log.Fatalf("Registry file Ill-formed (%s)", registryFilePath)
	}
	registry := make(map[string][]string)
	for _, handle := range handles {
		registry[handle.Handle] = handle.Pkgs
	}
	return registry
}