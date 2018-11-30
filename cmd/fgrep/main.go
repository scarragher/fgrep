package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
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
	showContent   *bool
	fileCount     int32
	matches       int32
	skipped       int32
	maxFileSize   *int64
)

const (
	maxFilesPerWorker = 45
)

func main() {

	inputDirectory := flag.String("i", "", "The directory to search")
	fileName := flag.String("f", "", "The filename to search for")
	verboseOutput = flag.Bool("v", false, "verbose output")
	content := flag.String("c", "", "the content to search for within files")
	workers := flag.Int("w", 4, "the amount of workers to utilise")
	maxFileSize = flag.Int64("fs", 2000, "the maximum file size to search (KB)")
	showContent = flag.Bool("content", false, "show the content instead of the filename")

	flag.Parse()

	if *fileName == "" && *content == "" {
		flag.PrintDefaults()
		return
	}

	if *inputDirectory == "" {
		cd := filepath.Dir(os.Args[0])
		log("Defaulting directory to %s", cd)
		*inputDirectory = cd
	}

	log("Searching %s for filenames like '%s' with content '%s' less than or equal to %dKB.", *inputDirectory, *fileName, *content, *maxFileSize)

	initWorkers(*workers)

	startTime := time.Now()

	files, err := ioutil.ReadDir(*inputDirectory)

	if err != nil {
		fmt.Printf("Failed to read directory %s, %v", *inputDirectory, err.Error())
		return
	}

	search(*inputDirectory, files, *fileName, *content, "main")

	waitGroup.Wait()

	timeTaken := time.Since(startTime)

	log("Searched %d/%d files, skipped %d files. Found %d matches in %f seconds", (fileCount - skipped), fileCount, skipped, matches, timeTaken.Seconds())
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

func search(directory string, files []os.FileInfo, filename string, content string, who string) {

	fileLen := len(files)
	halfFileLen := (fileLen / 2)
	anchor := fileLen

	if halfFileLen > maxFilesPerWorker {
		// try to get a free worker and have it pick up half the work
		childWorker, ok := workerQueue.Dequeue()
		if ok {
			anchor = halfFileLen
			workerFiles := files[anchor:]

			log("%s contains %d files, %d-%d will be processed by Worker[%d]", directory, fileLen, anchor, fileLen, childWorker.ID)

			startWorker(childWorker,
				func() {
					search(directory, workerFiles, filename, content, fmt.Sprintf("Worker[%d]", childWorker.ID))
				},
			)
		}
	}

	log("\t[%s]: scanning '%s' (%d-%d)", who, directory, 0, anchor)

	for _, file := range files[0:anchor] {
		fp := filepath.Join(directory, file.Name())

		if file.IsDir() {

			worker, ok := workerQueue.Dequeue()
			childFilePath := filepath.Join(directory, file.Name())
			childFiles, err := ioutil.ReadDir(childFilePath)

			if err != nil {
				log("Failed to read directory %s: %v", childFilePath, err.Error())
				continue
			}

			if ok {
				// a worker is available - send it to work
				workFunc := func() {
					search(childFilePath, childFiles, filename, content, fmt.Sprintf("Worker[%d]", worker.ID))
				}

				log("[%s]: Offloading work for '%s' to worker %d", who, fp, worker.ID)

				startWorker(worker, workFunc)
				continue
			} else {
				// no worker available, handle this synchronously
				search(childFilePath, childFiles, filename, content, who)
			}
			continue
		}

		atomic.AddInt32(&fileCount, 1)

		log("%s scanning file %s", who, file.Name())

		if filename != "" {
			if strings.Contains(file.Name(), filename) {
				fmt.Println(fp)
			} else {
				continue
			}
		}

		if content != "" {
			contentBytes := []byte(content)
			contentSize := int64(len(contentBytes))
			fileSize := file.Size()

			if (fileSize / 1024) > *maxFileSize {
				log("Skipped file %s, size was %d which is greater than max: %d", fp, fileSize, *maxFileSize)
				atomic.AddInt32(&skipped, 1)
				continue
			}
			if fileSize < contentSize {
				log("Skipped file %s, size was %d, wanted > %d", fp, fileSize, contentSize)
				atomic.AddInt32(&skipped, 1)
				continue
			}

			fileData, err := os.Open(fp)
			if err != nil {
				log("Failed to open %s. Error: %v", filename, err)
				continue
			}
			defer fileData.Close()

			fileContent, err := ioutil.ReadAll(fileData)
			if err != nil {
				log("Failed to read %s. Error: %v", filename, err)
				continue
			}

			found, ok := fgrep.Scan(content, path.Ext(fp), fileContent, *showContent)

			if !ok {
				log("skipped %s", filename)
				atomic.AddInt32(&skipped, 1)
				continue
			}
			if found {
				if !*showContent {
					fmt.Println(fp)
				}
				atomic.AddInt32(&matches, 1)
			}
		}
	}

	return
}

func log(text string, values ...interface{}) {
	if !*verboseOutput {
		return
	}

	fmt.Printf(text+"\n", values...)
}

func startWorker(worker *fgrep.Worker, workFunc func()) {
	mutex.Lock()
	waitGroup.Add(1)
	mutex.Unlock()

	worker.WorkFunc = workFunc
	completeFunc := func() {
		log("Worker %d finished\n", worker.ID)
		enqueueWorker(worker)
		waitGroup.Done()
	}

	go worker.DoWork(completeFunc)

}

func enqueueWorker(worker *fgrep.Worker) {
	workerQueue.Enqueue(worker)
}
