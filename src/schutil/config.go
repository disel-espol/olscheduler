package schutil

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/url"
	"strconv"

	"../worker"
)

type Config struct {
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

func LoadConfigFromFile(configFilepath string) Config {
	var config Config
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

func CreateWorkersArray(configFilepath string, config Config) []*worker.Worker {
	var workers []*worker.Worker
	for i := 0; i < len(config.Workers); i = i + 2 {
		// Make Workers with their Reverse Proxy Handlers
		weight, err := strconv.Atoi(config.Workers[i+1])
		if err != nil || weight < 0 {
			log.Fatalf("Config file Ill-formed (%s), every worker weight must be a positive number", configFilepath)
		}
		u, _ := url.Parse("http://" + config.Workers[i])
		proxy := worker.NewHTTPReverseProxy(u)
		workerConfig := worker.WorkerConfig{u, weight}
		workers = append(workers, worker.NewWorker(workerConfig, proxy))
	}

	return workers
}

func CreateRegistryFromFile(registryFilePath string) map[string][]string {
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
