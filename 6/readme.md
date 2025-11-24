# Exercise 6 (needs more practice)

`LPUSH queue job1` -> pushes a job to the left of the q (q head)
`BRPOP queue 0` -> blocks until a job appears, then pops the last element(q tail)

- server should add jobs to the queue, and worker should retrieve jobs

- Let's refactor 5/ to use Redis instead of in memory list in the server.
  This way: - job doesn't vanish if server restarts - mulitple workers can share the same q safely - workers don't have to poll the /next_job endpoint on hte server - there is no manual locking - it is persistent - workers wakeup instatnly as new jobs appear

Set up redis either using docker `docker run -d -p 6379:6379 --name redis redis:7` or natively.

- Update: added a client, and a websocket connection between client and server, so no shortpolling for job status. So worker now adds jobs to a channel, there is a go routine that is listening to this channel, when it received a job update, and broadcasts to the connections listening on that jobID

There is also a keep alive function that keeps connections alive, new go routine is spawned for this.

The websocket part was mostly taken from GPT, do need to go over it more carefully.

on macos:

```bash
brew install redis
brew services start redis

redis-cli ping // (yay i love it when you check for heartbeat using ping pong)
```

```
% curl -X POST http://localhost:8080/submit \
     -H "Content-Type: application/json" \
     -d '{"input": "hello world"}'

{"id":"20251123191517.511054","input":"hello world","status":"queued","result":"","created":"2025-11-23T19:15:17.511067+05:30"}
% curl "http://localhost:8080/status?id=20251123191517.511054"

{"status":"queued"}

% curl "http://localhost:8080/status?id=20251123191517.511054"

{"result":"HELLO WORLD","status":"done"}
```

This way the server and worker can go down, and come back up, and our results will persist in redis.

We use redis for

- in memory job q, once worker pops element using brpop, the job json is gone, so at most once guarantee,
- we also use it as key value datastore, job:<id> is key, and for the value we store status and result.

- ps using the word persists isn't accurate i realize, because redis is in memory.

```
redis-server
do ctrl+c here
...
4816:M 23 Nov 2025 19:25:41.990 * Saving the final RDB snapshot before exiting.
14816:M 23 Nov 2025 19:25:41.991 * BGSAVE done, 0 keys saved, 0 keys skipped, 88 bytes written.
14816:M 23 Nov 2025 19:25:41.998 * DB saved on disk
14816:M 23 Nov 2025 19:25:41.998 # Redis is now ready to exit, bye bye...
(base)
```

```
redis-cli shutdown
ps -o pid,%cpu,%mem,command | grep redis
```

```
to properly stop as it might be running as homebrew event
brew services stop redis
brew services start redis
```

## SSE

- HTTP 1.0 open,close, open close ...
- HTTP 1.1 Keep alive header. Keep the connection open. close
- Web sockets, update HTTP 1.1, so you switch protocols, bi directions connections. No order, sever can send, client can send, protocol by itself

Server sent event?

- Unidirectional, server just sending stuff. Get text/event-stream, server says ok, text/event-stream content type,
  so useful for streaming, logging etc.

Honestly can just use web sockets

## also

- go doesn't have a set type, so instead you do map[things i want a set of]struct{}

## Running This

```
// start the server, it will log that we connected to redis
cd server && go run main.go
// npm init -y && npm install ws node-fetch might be needed
cd client && node client.js
// job will be queued and there is a connection between client and server
cd worker && python worker.py
// now the worker will complete job, server is listening on hte channel, so it received the jobs, client received the job then ws connection is closed

go test -v // also we arent checking if job id for websocket is valid so it will work
```

## go notes

```go
package main

import (
	"fmt"
	"time"
)

// ----------------------
// CONCURRENCY BASICS
// ----------------------

func worker(id int, ch chan string) {
	// run as a goroutine; blocks until message arrives on channel
	for msg := range ch {
		fmt.Println("worker", id, "got:", msg)
	}
}

func main() {
	ch := make(chan string) // create channel for string messages

	// spawn two concurrent workers
	go worker(1, ch)
	go worker(2, ch)

	// send messages into channel
	ch <- "hello"
	ch <- "world"

	close(ch) // closes channel; range loop exits in workers

	// give workers time to finish
	time.Sleep(time.Second)
}

// ----------------------
// TIMERS AND TICKERS
// ----------------------

func timers() {
	timer := time.NewTimer(2 * time.Second)
	<-timer.C // blocks until 2s passed
	fmt.Println("Timer fired!")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for i := 0; i < 3; i++ {
		<-ticker.C
		fmt.Println("Tick", i)
	}
}

// ----------------------
// SELECT AND TIMEOUTS
// ----------------------

func timeoutExample() {
	ch := make(chan string)
	go func() {
		time.Sleep(2 * time.Second)
		ch <- "done"
	}()

	select {
	case msg := <-ch:
		fmt.Println("Received:", msg)
	case <-time.After(1 * time.Second):
		fmt.Println("Timeout!")
	}
}

// ----------------------
// DEADLINES AND KEEPALIVE
// ----------------------

// In WebSocket code:
// conn.SetReadDeadline(time.Now().Add(60 * time.Second))
// conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
// conn.SetPongHandler(func(string) error { ... }) // resets read deadline
// ticker := time.NewTicker(30 * time.Second)
// select { case <-ticker.C: conn.WriteMessage(websocket.PingMessage,nil) }

```

## key words to review

web sockets, pointer receiver, go routines, channels, timers, tickers, lpush, brpop, sync, redis, upgrader, sync.RWMutex, how there are not sets so map[string]struct{}, mu.Lock(), defer hub.mu.Unlock(), delete(set, c), len(set), Unmrshal and marshal (bytes <-> go type), type smt struct {}, rdb.HSet(), rdb.Subscribe("channel name"), struct/class for the hub ie job id to set of client connections mapping, to abstract away locking it, check origin, HGetAll
json.loads(job_data),r.publish("where", payload),

## misc

```
ps aux | grep worker.py
pkill -f worker.py
ps -o pid,ppid,pgid,command | grep worker.py

lsof -i :8080
kill -9 <pid>
```

Encountered this problem that run_all.sh, the ctrl+c wasn't getting propograted and the worker processes were still living, so there were so many workers, and as i made changes to workers, i would see still some of the jobs being handled the old way.
