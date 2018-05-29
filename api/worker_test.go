package fgrep

import "testing"

func TestDoWork(t *testing.T) {
	worker := Worker{ID: 1}

	worker.WorkFunc = func() {
		if worker.Status != StatusWorking {
			t.Errorf("Expected %s while working, got %s", StatusWorking, worker.Status)
		}
	}

	worker.DoWork(func() {})

	if worker.Status != StatusIdle {
		t.Errorf("Expected status %s when finished, got %s", StatusIdle, worker.Status)
	}
}
