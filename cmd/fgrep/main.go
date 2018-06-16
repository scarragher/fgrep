package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
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
	matches       int32
	skipped       int32
	maxFileSize   *int64
)

func main() {

	inputDirectory := flag.String("i", "", "The directory to search")
	fileName := flag.String("f", "", "The filename to search for")
	verboseOutput = flag.Bool("v", false, "verbose output")
	content := flag.String("c", "", "the content to search for within files")
	workers := flag.Int("w", 4, "the amount of workers to utilise")
	maxFileSize = flag.Int64("fs", 2000, "the maximum file size to search (KB)")

	flag.Parse()

	if *inputDirectory == "" {
		flag.PrintDefaults()
		return
	}

	if *fileName == "" && *content == "" {
		fmt.Println("No search criteria specified, specify either -f <filename> or -c <content>")
		return
	}

	fmt.Printf("Searching %s for filenames like '%s' with content '%s' less than or equal to %dKB.\n", *inputDirectory, *fileName, *content, *maxFileSize)

	initWorkers(*workers)

	startTime := time.Now()

	search(*inputDirectory, *fileName, *content, "main")

	waitGroup.Wait()

	timeTaken := time.Since(startTime)

	fmt.Printf("Searched %d/%d files, skipped %d files. Found %d matches in %f seconds", (fileCount - skipped), fileCount, skipped, matches, timeTaken.Seconds())
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

	log(fmt.Sprintf("\t[%s]: scanning '%s'\n", who, directory))

	files, err := ioutil.ReadDir(directory)

	if err != nil {
		log(fmt.Sprintf("Error scanning '%s': %s", directory, err.Error()))
		return
	}

	for _, file := range files {
		filepath := path.Join(directory, file.Name())

		if file.IsDir() {

			worker, ok := workerQueue.Dequeue()

			if ok {
				// a worker is available - send it to work
				mutex.Lock()
				waitGroup.Add(1)
				mutex.Unlock()

				worker.WorkFunc = func() {
					search(filepath, filename, content, fmt.Sprintf("Worker[%d]", worker.ID))
				}

				workCompleteFunc := func() {
					log(fmt.Sprintf("Worker %s finished\n", who))
					enqueueWorker(worker)
					waitGroup.Done()
				}

				log(fmt.Sprintf("[%s]: Offloading work for '%s' to worker %d\n", who, filepath, worker.ID))

				go worker.DoWork(workCompleteFunc)

				continue
			} else {
				// no worker available, handle this synchronously
				search(filepath, filename, content, who)
			}
			continue
		}

		atomic.AddInt32(&fileCount, 1)

		if filename != "" {
			if strings.Contains(file.Name(), filename) {
				fmt.Printf("Found %s/%s, %s\n", directory, file.Name(), who)
			} else {
				continue
			}
		}

		if content != "" {
			contentBytes := []byte(content)
			contentSize := int64(len(contentBytes))
			fileSize := file.Size()

			if (fileSize / 1024) > *maxFileSize {
				log(fmt.Sprintf("Skipped file %s, size was %d which is greater than max: %d\n", filepath, fileSize, *maxFileSize))
				atomic.AddInt32(&skipped, 1)
				continue
			}
			if fileSize < contentSize {
				log(fmt.Sprintf("Skipped file %s, size was %d, wanted > %d\n", filepath, fileSize, contentSize))
				atomic.AddInt32(&skipped, 1)
				continue
			}

			fileData, err := os.Open(filepath)
			if err != nil {
				log(fmt.Sprintf("Failed to open %s. Error: %v", filename, err))
				continue
			}
			defer fileData.Close()

			fileContent, err := ioutil.ReadAll(fileData)
			if err != nil {
				log(fmt.Sprintf("Failed to read %s. Error: %v", filename, err))
				continue
			}

			found, ok := fgrep.Scan(content, path.Ext(filename), fileContent)

			if !ok {
				log(fmt.Sprintf("skipped %s", filename))
				atomic.AddInt32(&skipped, 1)
				continue
			}
			if found {
				fmt.Printf("Found content in %s\\%s\n", directory, file.Name())
				atomic.AddInt32(&matches, 1)
			}
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
