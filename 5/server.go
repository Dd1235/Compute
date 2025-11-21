package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Job struct {
	ID      string    `json:"id"`
	Input   string    `json:"input"`
	Status  string    `json:"status"` // queued, processing, done
	Result  string    `json:"result"`
	Created time.Time `json:"created"`
}

var (
	jobQ       = make([]Job, 0)
	jobStore   = make(map[string]*Job)
	stateMutex = sync.Mutex{}
)

func handleSubmit(w http.ResponseWriter, r *http.Request) {
	type Req struct {
		Input string `json:"input"`
	}
	var body Req
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if body.Input == "" {
		http.Error(w, "Input cannot be empty", http.StatusBadRequest)
		return
	}

	job := Job{
		ID:      time.Now().Format("20060102150405.000000"),
		Input:   body.Input,
		Status:  "queued",
		Created: time.Now(),
	}

	fmt.Printf("handleSubmit hit! Job created %v\n", job)
	stateMutex.Lock()
	jobQ = append(jobQ, job)
	jobStore[job.ID] = &job
	stateMutex.Unlock()

	json.NewEncoder(w).Encode(job)
}

func handleNextJob(w http.ResponseWriter, r *http.Request) {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	if len(jobQ) == 0 {
		w.WriteHeader(204) // no content
		return
	}

	job := jobQ[0]
	jobQ = jobQ[1:]
	job.Status = "processing"
	jobStore[job.ID] = &job
	json.NewEncoder(w).Encode(job)
}

func handleJobResult(w http.ResponseWriter, r *http.Request) {
	type Req struct {
		ID     string `json:"id"`
		Result string `json:"result"`
	}
	var body Req
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Couldn't parse your json!", http.StatusBadRequest)
		return
	}
	stateMutex.Lock()
	defer stateMutex.Unlock()

	if job, ok := jobStore[body.ID]; ok {
		job.Status = "done"
		job.Result = body.Result
		json.NewEncoder(w).Encode(job)
		return
	}

	http.Error(w, "job not found! How did you even process it and send??", 404)
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	stateMutex.Lock()
	defer stateMutex.Unlock()

	job, ok := jobStore[id]
	if !ok {
		http.Error(w, "Job ID not found!", 404)
		return
	}

	json.NewEncoder(w).Encode(job)
}

func main() {
	fmt.Println("Welcome to my server!")
	log.Println("Server running at :8080")

	http.HandleFunc("/submit", handleSubmit)
	http.HandleFunc("/next_job", handleNextJob)
	http.HandleFunc("/job_result", handleJobResult)
	http.HandleFunc("/status", handleStatus)

	http.ListenAndServe(":8080", nil)
}

// curl -X POST http://localhost:8080/submit \
// -H "Content-Type: application/json" \
// -d '{"input": "hello world"}'

// curl "http://localhost:8080/status?id=20251121211520.392071"

// curl -X POST http://localhost:8080/job_result -H "Content-Type: application/json" -d '{"id":"20251121211520.392071", "result" : "sha256 of hello world"}'
