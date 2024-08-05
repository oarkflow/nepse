package csv

import (
	"bytes"
	"encoding/csv"
	"sync"

	"github.com/oarkflow/pkg/str"
)

func WriteWorker(data <-chan map[string]interface{}, wg *sync.WaitGroup, results chan<- map[string]interface{}) {
	defer wg.Done()
	for j := range data {
		results <- j
	}
}

func ProcessWrite(data []map[string]interface{}, noOfWorkers ...int) *bytes.Buffer {
	workers := 2
	if len(noOfWorkers) > 0 {
		workers = noOfWorkers[0]
	}
	byteBuffer := &bytes.Buffer{}
	writer := csv.NewWriter(byteBuffer)
	colPosition := make(map[int]string)
	count := 0
	for key, _ := range data[0] {
		colPosition[count] = key
		count++
	}

	jobs := make(chan map[string]interface{})
	results := make(chan map[string]interface{})

	// I think we need a wait group, not sure.
	wg := new(sync.WaitGroup)
	// start up some workers that will block and wait?
	for w := 1; w <= workers; w++ {
		wg.Add(1)
		go WriteWorker(jobs, wg, results)
	}

	// Go over a file line by line and queue up a ton of work
	go func() {
		for _, val := range data {
			jobs <- val
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	header := make(map[int]string)

	counter := 0
	for v := range results {
		h := make([]string, len(v))
		d := make([]string, len(v))
		idx := 0
		if counter == 0 {
			for col, _ := range v {
				header[idx] = col
				h[idx] = col
				idx++
			}
			err := writer.Write(h)
			if err != nil {
				panic(err)
			}
		}

		for col, val := range v {
			for id, head := range header {
				if col == head {
					d[id] = str.ToString(val)
				}
			}
		}

		err := writer.Write(d)
		if err != nil {
			panic(err)
		}
		counter++
	}
	writer.Flush()
	return byteBuffer
}
