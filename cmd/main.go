package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"sync"

	fgrep "github.com/scarragher/fgrep/api"
)

var (
	workerQueue fgrep.WorkerQueue
	waitGroup   sync.WaitGroup
)

func main() {

	inputDirectory := flag.String("i", "", "The directory to search")
	searchCriteria := flag.String("s", "", "The string to search for")

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

	initWorkers(4)
	search(*inputDirectory, *searchCriteria, "main")

	waitGroup.Wait()
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

	//fmt.Printf("[%s]: scanning '%s'\n", who, directory)
	files, err := ioutil.ReadDir(directory)

	if err != nil {
		fmt.Printf("Error scanning '%s': %s", directory, err.Error())
		return
	}

	for _, file := range files {
		if file.IsDir() {
			filepath := path.Join(directory, file.Name())
			worker, ok := workerQueue.Dequeue()
			if ok {
				worker.WorkFunc = func() {
					waitGroup.Add(1)
					search(filepath, criteria, fmt.Sprintf("Worker[%d]", worker.ID))
				}

				//fmt.Printf("[%s]: Offloading work for '%s' to worker %d\n", who, filepath, worker.ID)

				go worker.DoWork(func() {
					enqueueWorker(worker)
					waitGroup.Done()
				})

				continue
			} else {
				search(filepath, criteria, who)
			}

		}

		if strings.Contains(file.Name(), criteria) {
			fmt.Println("Found "+file.Name(), who)
		}
		fmt.Printf("[%s]: checking '%s'\n", who, file.Name())
	}

	return
}

func enqueueWorker(worker *fgrep.Worker) {
	workerQueue.Enqueue(worker)
}
