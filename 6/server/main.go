package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

// this just initializes the client, it doesn't actually connect to redis yet
// when you make your first command, it opens a TCP connection, and the pool reuses those connections
var rdb = redis.NewClient(&redis.Options{Addr: "localhost:6379"})

type Job struct {
	ID      string    `json:"id"`
	Input   string    `json:"input"`
	Status  string    `json:"status"`
	Result  string    `json:"result"`
	Created time.Time `json:"created"`
}

func handleSubmit(w http.ResponseWriter, r *http.Request) {
	type Req struct {
		Input string `json:"input"`
	}
	var body Req
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON, couldn't parse the body", http.StatusBadRequest)
		return
	}

	job := Job{
		ID:      time.Now().Format("20060102150405.000000"),
		Input:   body.Input,
		Status:  "queued",
		Created: time.Now(),
	}

	jobBytes, _ := json.Marshal(job)

	// we automatically create the key jobs if it doesn't exist
	err := rdb.LPush(ctx, "jobs", jobBytes).Err()
	if err != nil {
		http.Error(w, "failed to push the job to the q!", 500)
		return
	}
	// save metadata for status tracking
	// this uses redis hash, map kinda thing
	rdb.HSet(ctx, "job:"+job.ID, "status", "queued")
	json.NewEncoder(w).Encode(job)
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing ID", 400)
		return
	}
	// this gets all teh fields and values from the jobs hash
	data, err := rdb.HGetAll(ctx, "job:"+id).Result()
	if err != nil || len(data) == 0 {
		http.Error(w, "Job not found in q! what is this job id??", 404)
		return
	}
	json.NewEncoder(w).Encode(data)
}

func main() {
	fmt.Println("Server running on :8080")

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Count not connect to Redis: %v", err)
	} else {
		log.Println("Ey Pinged Redis got Pong!")
	}

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("pong"))
	})

	http.HandleFunc("/submit", handleSubmit)
	http.HandleFunc("/status", handleStatus)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
