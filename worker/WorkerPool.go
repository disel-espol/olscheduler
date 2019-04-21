package worker

import "net/url"

type WorkerPool struct {
	workers []*Worker
	channel chan msg
}

func (p *WorkerPool) GetTotalWorkers() int {
	return len(p.workers)
}

func (p *WorkerPool) getWorkerAt(pos int) *Worker {
	return p.workers[pos]
}

func (p *WorkerPool) GetWorkers() []*Worker {
	return p.workers
}

func (p *WorkerPool) AddWorker(workerUrl string, weight int) {
	u, _ := url.Parse("http://" + workerUrl)
	p.workers = append(p.workers, NewWorker(u, weight))
}

func (p *WorkerPool) FindWorker(workerUrl string) int {
	totalWorkers := len(p.workers)
	for i := 0; i < totalWorkers; i++ {
		if p.workers[i].GetURL() == workerUrl {
			return i
		}
	}
	return -1
}

func (p *WorkerPool) RemoveWorkerAt(target int) {
	// order matters. It could be faster if order doesn't matter
	// https://stackoverflow.com/a/37335777/5207721
	p.workers = append(p.workers[:target], p.workers[target+1:]...)
}

type msg struct {
	worker    *Worker
	isWorking bool
}

func (p *WorkerPool) UpdateWorkerLoad(worker *Worker, isWorking bool) {
	p.channel <- msg{worker, isWorking}
}

// ManagePool is meant to be run in a new goroutine. It runs forever,
// listening for messages in the worker pool channel and updating the
// load value in each worker.
// Updating the worker's load should be done exclusively in a dedicated
// goroutine to avoid race conditions.
func (p *WorkerPool) ManagePool() {
	for true {
		msg := <-p.channel
		if msg.isWorking {
			msg.worker.load++
		} else {
			msg.worker.load--
		}
	}
}

// NewWorkerPool is a public constructor for WorkerPool type.
func NewWorkerPool(workers []*Worker) WorkerPool {
	return WorkerPool{workers, make(chan msg)}
}
