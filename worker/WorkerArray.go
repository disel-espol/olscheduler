package worker

import "net/url"

func AddWorkerToArray(workers []*Worker, workerUrl string, weight int) []*Worker {
	u, _ := url.Parse("http://" + workerUrl)
	proxy := NewHTTPReverseProxy(u)
	workerConfig := WorkerConfig{u, weight}
	return append(workers, NewWorker(workerConfig, proxy))
}

func FindWorkerInArray(workers []*Worker, workerUrl string) int {
	for i := 0; i < len(workers); i++ {
		if workers[i].GetURL() == workerUrl {
			return i
		}
	}
	return -1
}

func RemoveWorkerFromArray(workers []*Worker, target int) []*Worker {
	// order matters. It could be faster if order doesn't matter
	// https://stackoverflow.com/a/37335777/5207721
	return append(workers[:target], workers[target+1:]...)
}
