package balancer

import (
	"net/http"
	"net/url"
	"sync"

	"github.com/disel-espol/olscheduler/httputil"
	"github.com/disel-espol/olscheduler/lambda"
)

type LeastLoaded struct {
	workerUrls []url.URL
	loadMap    map[url.URL]uint
	mutex      *sync.Mutex
}

func (b *LeastLoaded) getCurrentWorkerLoad(workerUrl url.URL) uint {
	workerLoad, _ := b.loadMap[workerUrl]
	return workerLoad
}

func (b *LeastLoaded) incrementWorkerLoad(workerUrl url.URL) {
	workerLoad, _ := b.loadMap[workerUrl]
	b.loadMap[workerUrl] = workerLoad + 1
}

func (b *LeastLoaded) decrementWorkerLoad(workerUrl url.URL) {
	workerLoad, _ := b.loadMap[workerUrl]
	b.loadMap[workerUrl] = workerLoad - 1
}

func (b *LeastLoaded) SelectWorker(r *http.Request, l *lambda.Lambda) (url.URL, *httputil.HttpError) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	workerUrls := b.workerUrls
	if len(workerUrls) == 0 {
		return url.URL{}, httputil.New500Error("Can't select worker, Workers empty")
	}

	leastLoadedUrl := workerUrls[0]
	lowestLoad := b.getCurrentWorkerLoad(leastLoadedUrl)
	for i := 1; i < len(workerUrls); i++ {
		tempUrl := workerUrls[i]
		tempLoad := b.getCurrentWorkerLoad(tempUrl)
		if tempLoad < lowestLoad {
			leastLoadedUrl = tempUrl
			lowestLoad = tempLoad
		}
	}

	b.incrementWorkerLoad(leastLoadedUrl)
	return leastLoadedUrl, nil
}

func (b *LeastLoaded) ReleaseWorker(workerURL url.URL) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.decrementWorkerLoad(workerURL)
}

func (b *LeastLoaded) AddWorker(workerURL url.URL) {
	b.workerUrls = append(b.workerUrls, workerURL)
}

func (b *LeastLoaded) GetAllWorkers() []url.URL {
	workerUrls := b.workerUrls

	dest := make([]url.URL, len(workerUrls))
	copy(dest, workerUrls)
	return dest
}

func (b *LeastLoaded) RemoveWorker(targetURL url.URL) {
	source := b.workerUrls
	targetIndex := findUrlInSlice(source, targetURL)
	b.workerUrls = append(source[:targetIndex], source[targetIndex+1:]...)
}

func NewLeastLoaded(workerUrls []url.URL) Balancer {
	return &LeastLoaded{workerUrls, make(map[url.URL]uint), &sync.Mutex{}}
}

func NewLeastLoadedFromJSONSlice(jsonSlice []string) Balancer {
	return NewLeastLoaded(createWorkerURLSlice(jsonSlice))
}
