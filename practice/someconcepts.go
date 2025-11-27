package main

import (
	"fmt"
	"strings"
)

const (
	Monday = iota + 1
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
	Sunday
)

const (
	Readable   = 1 << iota // 1 << 0 = 1
	Writable               // 010
	Executable             // 100
)

func main() {
	fmt.Println("Monday: ", Monday)
	fmt.Println("Tuesday: ", Tuesday)

	fmt.Printf("Executable: %03b\n", Executable)

	words := []string{}
	words = append(words, "hi ")
	words = append(words, "how ")
	words = append(words, "are ")
	words = append(words, "you ")

	chunk := words[0:2]

	line := strings.Join(words, "!") // put the second guy between each of the words
	fmt.Println(chunk)
	fmt.Println(line)
}
