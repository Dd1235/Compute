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
var rdb = redis.NewClient(&redis.Options{Addr: "localhost:6379"})

type Worker struct {
	ID       string `json:"id"`
	CPU      int    `json:"cpu"`
	MemoryGB int    `json:"memory_gb"`
}

type Job struct {
	ID     string `json:"id"`
	Input  string `json:"input"`
	Status string `json:"status"`
}

// add it to the workers set, and store the worker details as key value pairs
func handleRegister(w http.ResponseWriter, r *http.Request) {
	var worker Worker
	if err := json.NewDecoder(r.Body).Decode(&worker); err != nil {
		http.Error(w, "Couldn't get the worker fields from your request!", http.StatusBadRequest)
		return
	}
	// needed: check if worker already exists, or if its valid or something of that sort
	rdb.SAdd(ctx, "workers", worker.ID)
	rdb.HSet(ctx, "worker:"+worker.ID,
		"cpu", worker.CPU,
		"memory_gb", worker.MemoryGB,
		"status", "idle",
		"last_heartbeat", time.Now().Unix(),
	)
	log.Printf("Worker registered successfully: %v", worker.ID)
	w.WriteHeader(200)
}

// queue the job and put its details(only job id and status in the kv store) in the key value store
func handleSubmit(w http.ResponseWriter, r *http.Request) {
	var job Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, "Bad json body", http.StatusBadRequest)
		return
	}
	job.Status = "queued"
	jobJson, _ := json.Marshal(job)
	rdb.LPush(ctx, "jobs", jobJson)
	rdb.HSet(ctx, "job:"+job.ID, "status", "queued")
	log.Printf("Queued job: %v", job.ID)
	w.WriteHeader(200)
}

// deque a job using BRPop and assign it to a worker, so here coordinator is also responsible for assigning jobs, workers aren't picking up job
func scheduler() {
	for {
		jobData, err := rdb.BRPop(ctx, 0, "jobs").Result()
		if err != nil {
			continue
		}
		jobJson := jobData[1]
		var job Job
		if err := json.Unmarshal([]byte(jobJson), &job); err != nil {
			log.Printf("job corrupted! couldn't unmarshal it %v", err)
		}
		assigned := false

		workers, _ := rdb.SMembers(ctx, "workers").Result()
		for _, wid := range workers {
			status, _ := rdb.HGet(ctx, "worker:"+wid, "status").Result()
			if status == "idle" {
				assigned = true
				rdb.LPush(ctx, "worker:"+wid+":jobs", jobJson)
				rdb.HSet(ctx, "worker:"+wid, "status", "busy")
				rdb.HSet(ctx, "job:"+job.ID, "status", "assigned", "worker", wid)
				log.Printf("Assigned job %s to worker %s", job.ID, wid)
				break
			}
		}

		if !assigned {
			log.Printf("No idle workers, requeueing job %s", job.ID)
			rdb.RPush(ctx, "jobs", jobJson) // preserves FIFO order
			time.Sleep(2 * time.Second)
		}

	}
}

func main() {
	fmt.Println("Welcome to my server!")
	fmt.Println("Listening on :8080")

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Couldn't connect to redis instance! Error: %v", err)
	}

	go scheduler()

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong!"))
	})
	http.HandleFunc("/register", handleRegister)
	http.HandleFunc("/submit", handleSubmit)

	http.ListenAndServe(":8080", nil)
}
