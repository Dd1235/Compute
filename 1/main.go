package main

import (
	"fmt"
	"net/http"
)

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Someone hit this endpoint!")
	fmt.Fprintf(w, "hello!\n")
}

func headers(w http.ResponseWriter, r *http.Request) {
	for name, headers := range r.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong!\n")
}

func main() {
	fmt.Println("Hey there! Welcome to my web server. This is the first demonstration.")
	http.HandleFunc("/", hello)
	http.HandleFunc("/headers", headers)
	http.HandleFunc("/ping", handlePing)
	http.ListenAndServe(":8090", nil)
}

// go run main.go
// curl http://localhost:8090/
