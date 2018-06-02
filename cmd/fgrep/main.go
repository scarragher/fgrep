package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	fgrep "github.com/scarragher/fgrep/api"
)

var (
	workerQueue   fgrep.WorkerQueue
	waitGroup     sync.WaitGroup
	mutex         sync.Mutex
	verboseOutput *bool
	fileCount     int32
)

func main() {

	inputDirectory := flag.String("i", "", "The directory to search")
	fileName := flag.String("f", "", "The filename to search for")
	verboseOutput = flag.Bool("v", false, "verbose output")
	content := flag.String("c", "", "the content to search for within files")

	flag.Parse()

	if *inputDirectory == "" {
		flag.PrintDefaults()
		return
	}

	if *fileName == "" && *content == "" {
		fmt.Println("No search criteria specified, specify either -f <filename> or -c <content>")
		return
	}

	fmt.Printf("Searching %s for files like '%s'\n", *inputDirectory, *fileName)

	initWorkers(4)

	startTime := time.Now()

	search(*inputDirectory, *fileName, "", "main")

	waitGroup.Wait()
	timeTaken := time.Since(startTime)

	fmt.Printf("Searched %d files in %f seconds\n", fileCount, timeTaken.Seconds())
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

			mutex.Lock()
			worker, ok := workerQueue.Dequeue()
			mutex.Unlock()

			if ok {
				mutex.Lock()
				waitGroup.Add(1)
				mutex.Unlock()

				worker.WorkFunc = func() {
					search(filepath, filename, content, fmt.Sprintf("Worker[%d]", worker.ID))
				}

				log(fmt.Sprintf("[%s]: Offloading work for '%s' to worker %d\n", who, filepath, worker.ID))

				go worker.DoWork(func() {
					mutex.Lock()
					enqueueWorker(worker)

					waitGroup.Done()
					mutex.Unlock()
				})

				continue
			} else {
				search(filepath, filename, content, who)
			}
			continue
		}

		atomic.AddInt32(&fileCount, 1)

		log(fmt.Sprintf("Scanning %s %s\n", directory, file.Name()))

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
