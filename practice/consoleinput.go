package main

import (
	"bufio"
	"fmt"
	"os"
)

// build on underlying buffer
// .Scan(), buffer contents will be overwritten
// .Text() copy into new string

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Enter text: ")
		if !scanner.Scan() {
			// exit if input stream closed (EOF)
			break
		}
		input := scanner.Text()

		if input == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		fmt.Printf("you typed: %s\n", input)
	}

}
