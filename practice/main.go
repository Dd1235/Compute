package main

import (
	"fmt"
	"time"
)

func main() {
	m := map[string]int{
		"apple":  5,
		"banana": 2,
	}
	fmt.Println(m)

	for k, v := range m {
		fmt.Printf("%v - %v\n", k, v)
	}

	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(1 * time.Second)
		ch1 <- "msg from ch1"
	}()

	go func() {
		time.Sleep(4 * time.Second)
		ch2 <- "msg from ch2"
	}()

	for i := 0; i < 2; i++ {
		select {
		case msg := <-ch1:
			fmt.Println(msg)
		case msg := <-ch2:
			fmt.Println(msg)
		case <-time.After(2 * time.Second):
			fmt.Println("timeout!")
		}
	}

	// ch1 <- "adding something else to ch1" this won't work, there is no go routine waiting to receive from ch1, so this line blocks forever

	// can also have fire and forget channels, initialize like
	// make(char string, 1)

	go func() {
		ch1 <- "this go routine is not waiting for someone to receive"
	}()

	time.Sleep(2 * time.Second)

	select {
	case msg := <-ch1:
		fmt.Println("Received:", msg)
	default:
		fmt.Println("No message yet")
	}

}
