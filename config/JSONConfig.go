package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/disel-espol/olscheduler/proxy"
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
		Host:         c.Host,
		Port:         c.Port,
		Balancer:     createBalancerFromConfig(c),
		Registry:     createRegistryFromFile(c.Registry),
		ReverseProxy: proxy.NewHTTPReverseProxy(),
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
