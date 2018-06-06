package fgrep

import (
	"errors"
	"sync"
)

//WorkerQueue holds a queue of Workers and provides functionality to enqueue/dequeue a Worker
type WorkerQueue struct {
	sync.Mutex
	Max int

	workers []*Worker
}

//Dequeue dequeues a worker, returning a pointer to the worker a bool to indicate success
func (wq *WorkerQueue) Dequeue() (*Worker, bool) {

	wq.Lock()
	if len(wq.workers) < 1 {
		return nil, false
	}

	worker := wq.workers[0]
	wq.workers = wq.workers[1:]

	wq.Unlock()
	return worker, true
}

//Enqueue queues a worker and returns a bool to indicate success and an error if the enqueue operation was not successfull
func (wq *WorkerQueue) Enqueue(worker *Worker) (bool, error) {
	wq.Lock()
	if wq.Max != 0 && len(wq.workers) >= wq.Max {
		return false, errors.New("Maximum amount of workers reached")
	}

	wq.workers = append(wq.workers, worker)

	wq.Unlock()
	return true, nil
}
