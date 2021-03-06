package fgrep

//Status defines a Workers current status
type Status string

const (
	//StatusIdle indicates a worker is idle
	StatusIdle Status = "idle"

	//StatusWorking indicates a worker is currently working
	StatusWorking = "working"
)

//Worker wraps a task allowing it to be tracked and queued
type Worker struct {
	ID       int
	Status   Status
	WorkFunc func()
}

// WorkCompleteFunc executes when a worker has finished working
type WorkCompleteFunc func()

//DoWork starts a worker, changing the status, logging the time started/finished
func (worker *Worker) DoWork(onComplete WorkCompleteFunc) {

	worker.Status = StatusWorking

	worker.WorkFunc()

	worker.Status = StatusIdle

	onComplete()

}
