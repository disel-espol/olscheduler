package balancer

import (
	"log"
	"net/http"
	"net/url"

	"github.com/disel-espol/olscheduler/httputil"
	"github.com/disel-espol/olscheduler/lambda"

	"github.com/lafikl/consistent"
)

type ConsistentHashingBounded struct {
	hashRing      *consistent.Consistent
	workerNodeMap map[string]url.URL
}

func (b *ConsistentHashingBounded) SelectWorker(r *http.Request, l *lambda.Lambda) (url.URL, *httputil.HttpError) {
	if len(b.workerNodeMap) == 0 {
		return url.URL{}, httputil.New500Error("Can't select worker, Workers empty")
	}

	host, err := b.hashRing.GetLeast(r.URL.String())
	if err != nil {
		log.Fatal(err)
	}

	b.hashRing.Inc(host)
	return b.workerNodeMap[host], nil
}

func (b *ConsistentHashingBounded) ReleaseWorker(workerUrl url.URL) {
	b.hashRing.Done(workerUrl.String())
}

func (b ConsistentHashingBounded) AddWorker(workerUrl url.URL) {
	host := workerUrl.String()
	b.hashRing.Add(host)
	b.workerNodeMap[host] = workerUrl
}

func (b *ConsistentHashingBounded) RemoveWorker(workerUrl url.URL) {
	host := workerUrl.String()
	b.hashRing.Remove(host)
	delete(b.workerNodeMap, host)
}

func (b *ConsistentHashingBounded) GetAllWorkers() []url.URL {
	return getUrlsFromMap(b.workerNodeMap)
}

func NewConsistentHashingBounded(workerUrls []url.URL) Balancer {
	workerNodeMap := make(map[string]url.URL)
	for _, workerUrl := range workerUrls {
		workerNodeMap[workerUrl.String()] = workerUrl
	}

	return &ConsistentHashingBounded{consistent.New(), workerNodeMap}
}

func NewConsistentHashingBoundedFromJSONSlice(jsonSlice []string) Balancer {
	return NewConsistentHashingBounded(createWorkerURLSlice(jsonSlice))
}
