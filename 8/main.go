package main

import (
	"bufio"
	"fmt"
	"hash/fnv"
	"os"
	"strings"
	"sync"
)

// split file into roughly n equal chunks by lines
// so it will be split per line according to how your original file is
// so make sure words aren't being split over two different lines

// take file path, take number of chunks to make
// first get all the lines so lines = {line1, line2, line3}
// say chunk size = 2
// then first chunk get lines[0:2], second chunk lines[2]
func splitFile(path string, n int) []string {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	lines := []string{} // create an empty slice of strings

	scanner := bufio.NewScanner(file) // reads the file line by line

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) == 0 {
		return []string{}
	}

	chunkSize := (len(lines) + n - 1) / n // ceil
	chunks := []string{}
	for i := 0; i < len(lines); i += chunkSize {
		end := i + chunkSize
		if end > len(lines) {
			end = len(lines)
		}
		chunks = append(chunks, strings.Join(lines[i:end], "\n"))
	}

	return chunks
}

func partition(key string, numReducers int) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32()) % numReducers
}

type KeyValue struct {
	Key   string
	Value int
}

// groupByKey performs the Shuffle step
// [(apple 1) (orange 1) (apple 1)] [(apple 1)] intermediates
// (apple [1 1 1] orange [1])

func groupByKey(intermediates [][]KeyValue) map[string][]int {
	grouped := make(map[string][]int)
	for _, kvs := range intermediates {
		for _, kv := range kvs {
			grouped[kv.Key] = append(grouped[kv.Key], kv.Value)
		}
	}
	return grouped
}

type MapFunc func(string) []KeyValue
type ReduceFunc func(string, []int) KeyValue

type Job struct {
	InputFile   string
	Map         MapFunc
	Reduce      ReduceFunc
	NumMappers  int
	NumReducers int
}

type Master struct {
	Job           Job
	Intermediates [][]KeyValue
	Results       []KeyValue
}

func (m *Master) Run() {
	//Split input file into chunks
	chunks := splitFile(m.Job.InputFile, m.Job.NumMappers)

	// Map phase (each mapper runs concurrently)
	mapChan := make(chan []KeyValue, len(chunks))
	var wg sync.WaitGroup
	for _, chunk := range chunks {
		wg.Add(1)
		go func(data string) {
			defer wg.Done()
			fmt.Printf("Mapper is processing\n %v\n", data)
			kvs := m.Job.Map(data)
			mapChan <- kvs
		}(chunk)
	}
	wg.Wait()
	close(mapChan)

	for kvs := range mapChan {
		// one for each of the workers
		m.Intermediates = append(m.Intermediates, kvs)
	}

	//  Shuffle phase (group by key)
	grouped := groupByKey(m.Intermediates)

	// Reduce phase (parallel reducers)
	reduceChan := make(chan KeyValue, len(grouped))
	var rwg sync.WaitGroup

	numReducers := m.Job.NumReducers
	buckets := make([]map[string][]int, numReducers)
	for i := 0; i < numReducers; i++ {
		buckets[i] = make(map[string][]int)
	}

	// assign each key to a reducer bucket
	for key, vals := range grouped {
		idx := partition(key, numReducers)
		buckets[idx][key] = vals
	}

	// launch exactly NumReducers goroutines
	for i := 0; i < numReducers; i++ {
		rwg.Add(1)
		go func(bucket map[string][]int) {
			defer rwg.Done()
			for k, v := range bucket {
				reduceChan <- m.Job.Reduce(k, v)
			}
		}(buckets[i])
	}

	rwg.Wait()
	close(reduceChan)

	for kv := range reduceChan {
		m.Results = append(m.Results, kv)
	}

}

func WordCountMap(data string) []KeyValue {
	words := strings.Fields(data) //splits at ' ' '\t' '\n' etc.
	out := make([]KeyValue, 0, len(words))
	for _, w := range words {
		out = append(out, KeyValue{Key: w, Value: 1})
	}
	return out
}

func WordCountReduce(key string, values []int) KeyValue {
	sum := 0
	for _, v := range values {
		sum += v
	}
	return KeyValue{Key: key, Value: sum}
}

func main() {
	fmt.Println("Simplest map reduce!")

	job := Job{
		InputFile:   "./input.txt",
		Map:         WordCountMap,
		Reduce:      WordCountReduce,
		NumMappers:  2,
		NumReducers: 2,
	}

	master := Master{Job: job}

	fmt.Println("\n--- Computing..... ---")

	master.Run()

	fmt.Println("\n--- Final Results ---")
	for _, kv := range master.Results {
		fmt.Printf("%s: %d\n", kv.Key, kv.Value)
	}
}
