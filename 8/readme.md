Here we implement a very simple map reduce to count the frequency of words in an input file
We spin up go routines for mapper and reducer workers. And the shuffling part is technically done by coordinator so isn't "distributed", that does mean i haven't accomated for the case where entire data doesn't fit into memory

main abstractions:

```Go
type KeyValue struct {
	Key   string
	Value int
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

// pointer receiver method
func (m *Master) Run() {

    // here it is implement such that you split into lines, and lines you divide equally between the chunks
	chunks := splitFile(m.Job.InputFile, m.Job.NumMappers)

    // chunk = some number of lines being handled by one Mapper
    // so no(chunks) = no(mappers)
    // for each mapper we need one slice of keyvalue pairs

	mapChan := make(chan []KeyValue, len(chunks))
	var wg sync.WaitGroup // define wait group
	for _, chunk := range chunks {
		wg.Add(1)
		go func(data string) {
			defer wg.Done()
			fmt.Printf("Mapper is processing\n %v\n", data)
			kvs := m.Job.Map(data) // get back []KeyValue eg [(apple,1),(banana,1),(apple,1)]
			mapChan <- kvs // add this slice to channel
		}(chunk)
	}
	wg.Wait()
	close(mapChan)

	for kvs := range mapChan {
		// one for each of the workers
		m.Intermediates = append(m.Intermediates, kvs)
	}

	//  Shuffle phase (group by key)
    // so this one is just handled by the master, perhaps this can be parallelised or we can have each of the mappers also performing a reduce job on their data later, this should be okay as long as operation is ummm assosiative?
    // so [(apple,[1,1]), (banana,[1])]
	grouped := groupByKey(m.Intermediates)

	// Reduce phase (parallel reducers)
    // one for each type of key
	reduceChan := make(chan KeyValue, len(grouped))
	var rwg sync.WaitGroup

	numReducers := m.Job.NumReducers

    // make numReducers num of buckets, buckets[i] -> bucket of reducer i
    // bucket -> [(k1,[...]),]
	buckets := make([]map[string][]int, numReducers)
	for i := 0; i < numReducers; i++ {
		buckets[i] = make(map[string][]int)
	}

	// assign each key to a reducer bucket, use hashing for this
	for key, vals := range grouped {
		idx := partition(key, numReducers)
		buckets[idx][key] = vals
	}

	// launch exactly NumReducers goroutines, once you have used map and gotten (apple,1),and shuffling to just get all values related to one key in one place so (apple, [1,1]), just pass this to reduce
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

```

So the user just has to give give the map function and the reduce function.
