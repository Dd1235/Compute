package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

var ctx = context.Background()

// this just initializes the client, it doesn't actually connect to redis yet
// when you make your first command, it opens a TCP connection, and the pool reuses those connections
var rdb = redis.NewClient(&redis.Options{Addr: "localhost:6379"})

// allow all origins for now
//
//	var upgrader = websocket.Upgrader{
//		CheckOrigin: func(r *http.Request) bool { return true },
//	}
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")

		// Allow browser JS from localhost
		if origin == "http://localhost:8080" ||
			origin == "http://127.0.0.1:8080" ||
			origin == "http://localhost:3000" {
			return true
		}

		// Allow curl or Go scripts (they often send no Origin header)
		if origin == "" {
			return true
		}

		// add your production origin here
		return false
	},
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*websocket.Conn]struct{}
}

type Job struct {
	ID      string    `json:"id"`
	Input   string    `json:"input"`
	Status  string    `json:"status"`
	Result  string    `json:"result"`
	Created time.Time `json:"created"`
}

// var clients = make(map[string]*websocket.Conn)
// we want this data structure mapping the jobid to the connection to be safe

var hub = &Hub{clients: make(map[string]map[*websocket.Conn]struct{})}

func (h *Hub) add(jobID string, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[jobID] == nil {
		h.clients[jobID] = make(map[*websocket.Conn]struct{})
	}
	h.clients[jobID][c] = struct{}{}
}
func (h *Hub) remove(jobID string, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set, ok := h.clients[jobID]; ok {
		delete(set, c)
		if len(set) == 0 {
			delete(h.clients, jobID)
		}
	}
}

// send a message to all the connections subscribed to that jobID
func (h *Hub) broadcastAndClose(jobID string, msg []byte) {
	h.mu.RLock()
	conns := make([]*websocket.Conn, 0, len(h.clients[jobID]))
	for c := range h.clients[jobID] {
		conns = append(conns, c)
	}
	h.mu.RUnlock()

	for _, c := range conns {
		_ = c.WriteMessage(websocket.TextMessage, msg)
		_ = c.Close()
		h.remove(jobID, c)
	}
}

// compile time values, used as timeouts
// ping pong are control frames part of ws, not user mesages
const (
	wsReadWait  = 60 * time.Second
	wsPongWait  = 60 * time.Second // how long to wait for pong after a ping
	wsPingEvery = 30 * time.Second // interval between pings
	wsWriteWait = 10 * time.Second //maximum time we allow a write to take before aborting
)

func keepAlive(conn *websocket.Conn, jobID string) {
	conn.SetReadLimit(1 << 20)                           // incoming websocket can be a maximum of about 1 MB
	_ = conn.SetReadDeadline(time.Now().Add(wsReadWait)) // return error on read message is no message received before that time

	// call back function
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(wsPongWait))
		return nil
	})

	ticker := time.NewTicker(wsPingEvery)
	defer ticker.Stop()

	for {
		// read pump, is client disconeected?
		if _, _, err := conn.ReadMessage(); err != nil {
			hub.remove(jobID, conn)
			_ = conn.Close()
			return
		}

		// ping (optional: move ping to a separate goroutine if you expect messages in)
		// select is waiting on multiple events
		select {
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				hub.remove(jobID, conn)
				_ = conn.Close()
				return
			}
		default:
		}
	}
}

func handleWS(w http.ResponseWriter, r *http.Request) {
	jobID := r.URL.Query().Get("job_id")
	if jobID == "" {
		http.Error(w, "Missing job_id!", http.StatusBadRequest)
		return
	}
	// updgrade this http connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}
	// i guess we also need to check the jobID belongs to a previously submitted job!
	hub.add(jobID, conn)
	log.Printf("Client connected for job %s", jobID)
	go keepAlive(conn, jobID) // don’t block handler, so spawn a go routine for that one connection, main go routine returns allowing http server to accept more clients
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

	// adding a listener that forwards updates to clients

	type JobUpdate struct {
		JobID  string `json:"job_id"`
		Status string `json:"status"`
		Result string `json:"result"`
	}

	go func() {
		sub := rdb.Subscribe(ctx, "job_updates")
		ch := sub.Channel()
		for msg := range ch {
			var update JobUpdate
			if err := json.Unmarshal([]byte(msg.Payload), &update); err != nil {
				log.Printf("Invalid payload, couldn't marshall it into job update : %v", err)
				continue
			}
			data, err := json.Marshal(update)
			if err != nil {
				log.Printf("marshal error: %v", err)
				continue
			}
			hub.broadcastAndClose(update.JobID, data)
		}
	}()

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("pong"))
	})

	http.HandleFunc("/submit", handleSubmit)
	http.HandleFunc("/status", handleStatus)
	http.HandleFunc("/ws", handleWS)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
