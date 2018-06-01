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
	workerQueue   fgrep.WorkerQueue
	waitGroup     sync.WaitGroup
	mutex         sync.Mutex
	verboseOutput *bool
	fileCount     int
)

func main() {

	inputDirectory := flag.String("i", "", "The directory to search")
	fileName := flag.String("f", "", "The filename to search for")
	verboseOutput = flag.Bool("v", false, "verbose output")

	flag.Parse()

	if *inputDirectory == "" {
		fmt.Println("No directory specified")
		return
	}

	if *fileName == "" {
		fmt.Println("No search criteria specified")
		return
	}

	fmt.Printf("Searching %s for '%s'\n", *inputDirectory, *fileName)
	startTime := time.Now()

	initWorkers(4)
	search(*inputDirectory, *fileName, "", "main")

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

func search(directory string, filename string, content string, who string) {

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
					search(filepath, filename, content, fmt.Sprintf("Worker[%d]", worker.ID))
				}

				log(fmt.Sprintf("[%s]: Offloading work for '%s' to worker %d\n", who, filepath, worker.ID))

				go worker.DoWork(func() {
					enqueueWorker(worker)
					exclusive(func() { waitGroup.Done() })
				})

				continue
			} else {
				search(filepath, filename, content, who)
			}
			continue
		}

		mutex.Lock()
		fileCount++
		mutex.Unlock()

		if strings.Contains(file.Name(), filename) {
			fmt.Printf("Found %s/%s, %s\n", directory, file.Name(), who)
		}
	}

	return
}

func log(text string) {
	if !*verboseOutput {
		return
	}

	fmt.Print(text)
}

func enqueueWorker(worker *fgrep.Worker) {
	workerQueue.Enqueue(worker)
}

func exclusive(f func()) {
	mutex.Lock()
	f()
	mutex.Unlock()
}
