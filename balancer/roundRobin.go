package balancer

import (
	"net/http"
	"net/url"

	"github.com/disel-espol/olscheduler/httputil"
	"github.com/disel-espol/olscheduler/lambda"
	"github.com/disel-espol/olscheduler/thread"
)

type RoundRobin struct {
	counter    *thread.Counter
	workerUrls []url.URL
}

func (b *RoundRobin) SelectWorker(r *http.Request, l *lambda.Lambda) (url.URL, *httputil.HttpError) {
	workerUrls := b.workerUrls
	var totalWorkers uint = uint(len(workerUrls))
	if totalWorkers == 0 {
		return url.URL{}, httputil.New500Error("Can't select worker, Workers empty")
	}

	currentIndex := b.counter.Inc() % totalWorkers

	return workerUrls[currentIndex], nil
}

func (b *RoundRobin) ReleaseWorker(workerURL url.URL) {
}

func (b *RoundRobin) AddWorker(workerURL url.URL) {
	b.workerUrls = append(b.workerUrls, workerURL)
}

func (b *RoundRobin) RemoveWorker(targetURL url.URL) {
	source := b.workerUrls
	targetIndex := findUrlInSlice(source, targetURL)
	if targetIndex > -1 {
		b.workerUrls = append(source[:targetIndex], source[targetIndex+1:]...)
	}
}

func (b *RoundRobin) GetAllWorkers() []url.URL {
	workerUrls := b.workerUrls

	dest := make([]url.URL, len(workerUrls))
	copy(dest, workerUrls)
	return dest
}

func NewRoundRobin(workerUrls []url.URL) Balancer {
	return &RoundRobin{thread.NewCounter(), workerUrls}
}

func NewRoundRobinFromJSONSlice(jsonSlice []string) Balancer {
	return NewRoundRobin(createWorkerURLSlice(jsonSlice))
}
