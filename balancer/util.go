package balancer

import (
	"log"
	"net/url"
	"strconv"

	"github.com/disel-espol/olscheduler/balancer/worker"
)

func createWorkerURLSlice(jsonSlice []string) []url.URL {
	workersLength := len(jsonSlice)
	workerUrls := make([]url.URL, workersLength)

	for i, urlString := range jsonSlice {
		workerUrl, err := url.Parse(urlString)
		if err != nil {
			log.Fatalf("Config file Ill-formed, unable to parse URL " + urlString)
		}
		workerUrls[i] = *workerUrl
	}

	return workerUrls
}

func createWeightedNodeSlice(jsonSlice []string) []worker.WeightedNode {
	workersLength := len(jsonSlice)
	if workersLength%2 == 1 {
		log.Fatalf("Config file Ill-formed, every worker url must be followed by its weight")
	}

	workerSlice := make([]worker.WeightedNode, workersLength/2)
	for i := 0; i < workersLength; i = i + 2 {
		weight, err := strconv.Atoi(jsonSlice[i+1])
		if err != nil || weight < 0 {
			log.Fatalf("Config file Ill-formed, every worker weight must be a positive number")
		}
		urlString := "http://" + jsonSlice[i]
		workerUrl, err := url.Parse(urlString)
		if err != nil {
			log.Fatalf("Config file Ill-formed, unable to parse URL " + urlString)
		}
		workerSlice[i/2] = worker.NewWeightedNode(*workerUrl, weight)
	}

	return workerSlice
}

func findWeightedNodeInSlice(nodeSlice []worker.WeightedNode, targetUrl url.URL) int {
	totalItems := len(nodeSlice)
	for i := 0; i < totalItems; i++ {
		if nodeSlice[i].GetURL() == targetUrl {
			return i
		}
	}
	return -1
}

func findUrlInSlice(urlSlice []url.URL, target url.URL) int {
	totalItems := len(urlSlice)
	for i := 0; i < totalItems; i++ {
		if urlSlice[i] == target {
			return i
		}
	}
	return -1
}
func getUrlsFromMap(urlMap map[string]url.URL) []url.URL {
	totalUrls := len(urlMap)
	urlSlice := make([]url.URL, totalUrls)
	i := 0
	for _, indexedUrl := range urlMap {
		urlSlice[i] = indexedUrl
		i++
	}
	return urlSlice
}
