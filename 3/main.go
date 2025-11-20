package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
)

func handlePing(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "pong!\n")
}

func handleHome(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Hello, welcome to our server! It is written in Go!\n")
}

// for destructioning the json response
type Job struct {
	Name string
}

func handleSubmit(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("Submit endpoint hit!\n")
	var j Job

	// go handler is expecting explicit json
	// we only explicitly accept JSON that matches the Job
	// if you want curl -X POST -d "Name:Deepya" http://localhost:8090/submit  to work then, you need to parseForm fomr req, and just get the name field

	err := json.NewDecoder(req.Body).Decode(&j)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h := sha256.New()
	h.Write([]byte(j.Name))
	bs := h.Sum(nil)

	fmt.Fprintf(w, "Your name %v is hashed to: %x\n", j.Name, bs)
}

func main() {
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/ping", handlePing)
	http.HandleFunc("/submit", handleSubmit)
	http.ListenAndServe(":8090", nil)
}

// curl -X POST -H "Content-Type: application/json" \
//   -d '{"Name":"Deepya"}' \
//   localhost:8090/submit
