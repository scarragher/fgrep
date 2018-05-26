package fgrep

import "testing"

func TestEnqueue(t *testing.T) {
	wq := WorkerQueue{}

	worker1 := Worker{ID: 1}

	ok, err := wq.Enqueue(&worker1)

	if !ok {
		t.Error("Failed to queue worker")
	}

	if err != nil {
		t.Error(err.Error())
	}

	if len(wq.workers) != 1 {
		t.Errorf("Expected %d workers, got %d", 1, len(wq.workers))
	}
}

func TestDequeue(t *testing.T) {
	wq := WorkerQueue{}

	worker1 := Worker{ID: 1}
	worker2 := Worker{ID: 2}

	ok, _ := wq.Enqueue(&worker1)

	if !ok {
		t.Error("Failed to queue worker1")
	}

	ok, _ = wq.Enqueue(&worker2)

	if !ok {
		t.Error("Failed to queue worker2")
	}

	w1, ok := wq.Dequeue()
	if !ok {
		t.Error("Failed to dequeue worker 1")
	}

	if w1 != &worker1 {
		t.Errorf("Expected worker %p, got %v", &worker1, &w1)
	}

	if len(wq.workers) != 1 {
		t.Error("Expected 1 worker in queue")
	}

	w2, ok := wq.Dequeue()
	if !ok {
		t.Error("Failed to queue worker 2")
	}

	if w2 != &worker2 {
		t.Error("Expected worker %p, got %p", &worker2, &w2)
	}

	if len(wq.workers) != 0 {
		t.Error("Expected 0 workers in queue")
	}
}

func TestNoWorkers(t *testing.T) {
	wq := WorkerQueue{}

	worker, ok := wq.Dequeue()

	if ok {
		t.Error("Dequeue with no workers should not be ok")
	}

	if worker != nil {
		t.Error("Worker should be nil ")
	}
}
func TestMaxWorkers(t *testing.T) {
	wq := WorkerQueue{Max: 5}

	ok, err := wq.Enqueue(&Worker{ID: 1})

	if !ok && err != nil {
		t.Error("Expected worker 1 to be queued")
	}

	ok, err = wq.Enqueue(&Worker{ID: 2})

	if !ok && err != nil {
		t.Error("Expected worker 2 to be queued")
	}

	ok, err = wq.Enqueue(&Worker{ID: 3})

	if !ok && err != nil {
		t.Error("Expected worker 3 to be queued")
	}

	ok, err = wq.Enqueue(&Worker{ID: 4})

	if !ok && err != nil {
		t.Error("Expected worker 4 to be queued")
	}

	ok, err = wq.Enqueue(&Worker{ID: 5})

	if !ok && err != nil {
		t.Error("Expected worker 5 to be queued")
	}

	ok, err = wq.Enqueue(&Worker{ID: 6})

	if ok {
		t.Error("Worker 6 should not be ok")
	}

	if err == nil {
		t.Error("Worker 6 enqueue should have errored")
	}

}
