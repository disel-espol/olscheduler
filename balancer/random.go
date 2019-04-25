package balancer

import (
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/disel-espol/olscheduler/balancer/worker"
	"github.com/disel-espol/olscheduler/httputil"
	"github.com/disel-espol/olscheduler/lambda"
)

// Select a random worker
type Random struct {
	workerNodes []worker.WeightedNode
}

func calculateTotalWeight(workers []worker.WeightedNode) int {
	totalWeight := 0
	for i, _ := range workers {
		if workers[i].GetWeight() < 0 {
		}
		totalWeight += workers[i].GetWeight()
	}
	return totalWeight
}

func chooseFromWeightedRandomValue(workers []worker.WeightedNode, totalWeight int, randomValue int) (url.URL, *httputil.HttpError) {
	accumWeight := 0
	for i, _ := range workers {
		if workers[i].GetWeight() == 0 {
			continue
		}
		accumWeight += workers[i].GetWeight()
		if accumWeight >= randomValue {
			return workers[i].GetURL(), nil
		}
	}

	return url.URL{}, httputil.New500Error("Can't select worker, All weights are zero")
}

func (b *Random) SelectWorker(r *http.Request, l *lambda.Lambda) (url.URL, *httputil.HttpError) {
	workers := b.workerNodes
	totalWorkers := len(workers)
	if totalWorkers == 0 {
		return url.URL{}, httputil.New500Error("Can't select worker, Workers empty")
	}
	totalWeight := calculateTotalWeight(workers)
	weightedRandomValue := rand.Intn(totalWeight) + 1
	return chooseFromWeightedRandomValue(workers, totalWeight, weightedRandomValue)
}

func (b *Random) AddWorker(workerURL url.URL) {
	b.workerNodes = append(b.workerNodes, worker.NewWeightedNode(workerURL, 1))
}

func (b *Random) ReleaseWorker(workerURL url.URL) {
}

func (b *Random) RemoveWorker(targetURL url.URL) {
	source := b.workerNodes
	targetIndex := findWeightedNodeInSlice(source, targetURL)
	if targetIndex > -1 {
		b.workerNodes = append(source[:targetIndex], source[targetIndex+1:]...)
	}
}

func (b *Random) GetAllWorkers() []url.URL {
	workerNodes := b.workerNodes
	totalWorkers := len(workerNodes)
	workerUrls := make([]url.URL, totalWorkers)

	for i, indexedNode := range workerNodes {
		workerUrls[i] = indexedNode.GetURL()
	}
	return workerUrls
}

func NewRandom(workerNodes []worker.WeightedNode) Balancer {
	rand.Seed(time.Now().Unix()) // For rand future calls
	return &Random{workerNodes}
}

func NewRandomFromJSONSlice(jsonSlice []string) Balancer {
	return NewRandom(createWeightedNodeSlice(jsonSlice))
}
