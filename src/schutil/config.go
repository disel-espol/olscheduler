package schutil

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http/httputil"
	"net/url"
	"strconv"
)

type Config struct {
	Host          string   `json:"host"`
	Port          int      `json:"port"`
	Balancer      string   `json:"balancer"`
	LoadThreshold int      `json:"load-threshold"`
	Workers       []string `json:"workers"`
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

func CreateWorkersArray(configFilepath string, config Config) []Worker {
	var workers []Worker
	for i := 0; i < len(config.Workers); i = i + 2 {
		// Make Workers with their Reverse Proxy Handlers
		weight, err := strconv.Atoi(config.Workers[i+1])
		if err != nil || weight < 0 {
			log.Fatalf("Config file Ill-formed (%s), every worker weight must be a positive number", configFilepath)
		}
		u, _ := url.Parse("http://" + config.Workers[i])
		handler := httputil.NewSingleHostReverseProxy(u)
		workers = append(workers, Worker{u, handler, 0, weight})
	}

	return workers
}
