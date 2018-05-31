package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"sync"
	"time"

	fgrep "github.com/scarragher/fgrep/api"
)

var (
	workerQueue    fgrep.WorkerQueue
	waitGroup      sync.WaitGroup
	mutex          sync.Mutex
	loggingEnabled *bool
	fileCount      int
)

func main() {

	inputDirectory := flag.String("i", "", "The directory to search")
	searchCriteria := flag.String("s", "", "The string to search for")
	loggingEnabled = flag.Bool("l", false, "enable logging")

	flag.Parse()

	if *inputDirectory == "" {
		fmt.Println("No directory specified")
		return
	}

	if *searchCriteria == "" {
		fmt.Println("No search criteria specified")
		return
	}

	fmt.Printf("Searching %s for '%s'\n", *inputDirectory, *searchCriteria)
	startTime := time.Now()

	initWorkers(4)
	search(*inputDirectory, *searchCriteria, "main")

	waitGroup.Wait()
	timeTaken := time.Since(startTime)

	fmt.Printf("Searched %d files in %f seconds", fileCount, timeTaken.Seconds())
}

func initWorkers(maxWorkers int) {

	workerQueue = fgrep.WorkerQueue{
		Max: maxWorkers,
	}

	for i := 0; i < maxWorkers; i++ {
		w := fgrep.Worker{
			ID: i,
		}

		workerQueue.Enqueue(&w)
	}

}

func search(directory string, criteria string, who string) {

	log(fmt.Sprintf("[%s]: scanning '%s'\n", who, directory))
	files, err := ioutil.ReadDir(directory)

	if err != nil {
		log(fmt.Sprintf("Error scanning '%s': %s", directory, err.Error()))
		return
	}

	for _, file := range files {
		if file.IsDir() {
			filepath := path.Join(directory, file.Name())
			worker, ok := workerQueue.Dequeue()
			if ok {
				mutex.Lock()
				waitGroup.Add(1)
				mutex.Unlock()

				worker.WorkFunc = func() {
					search(filepath, criteria, fmt.Sprintf("Worker[%d]", worker.ID))
				}

				log(fmt.Sprintf("[%s]: Offloading work for '%s' to worker %d\n", who, filepath, worker.ID))

				go worker.DoWork(func() {
					log(fmt.Sprintf("Worker[%d]: Finished\n", worker.ID))
					enqueueWorker(worker)
					mutex.Lock()
					waitGroup.Done()
					mutex.Unlock()
				})

				continue
			} else {
				search(filepath, criteria, who)
			}
			continue
		}

		mutex.Lock()
		fileCount++
		mutex.Unlock()

		if strings.Contains(file.Name(), criteria) {
			log(fmt.Sprintf("Found %s/%s, %s\n", directory, file.Name(), who))
		}
	}

	if who == "main" {
		log("[Main] finished")
	}
	return
}

func log(text string) {
	if !*loggingEnabled {
		return
	}

	fmt.Print(text)
}

func enqueueWorker(worker *fgrep.Worker) {
	workerQueue.Enqueue(worker)
}
