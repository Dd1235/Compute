package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

func doWork(d time.Duration, s string, resch chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	time.Sleep(d)
	resch <- s
}

func main() {
	start := time.Now()
	wg := &sync.WaitGroup{}
	wg.Add(3)

	resultch := make(chan string)

	words := []string{}

	go doWork(2*time.Second, "work1 ", resultch, wg)
	go doWork(4*time.Second, "work2 ", resultch, wg)
	go doWork(6*time.Second, "work3 ", resultch, wg)

	go func() {

		for r := range resultch {
			words = append(words, r)
		}
		fmt.Printf("work took %v second\n", time.Since(start))

	}()

	wg.Wait()
	close(resultch)
	time.Sleep(3 * time.Second)
	res := strings.Join(words, "|")

	fmt.Printf("Combined res: %v\n", res)

}
