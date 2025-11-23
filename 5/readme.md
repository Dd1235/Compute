# Exercise 5: Worker uses short polling to pull next_job from server, client also checks /status endpoint for update

current model is pull based.

To eliminate polling, can have push model, but worker must run its own http server, have an open port, server needs to know workers ip/port

Push wastes less cycles but we must also keep the worker reachable.

Worker can be a server, doesn't waste cpu cycles

```python
from flask import Flask, request

app = Flask(__name__)

@app.post("/run_job")
def run_job():
    job = request.json
    # compute result
    return {"result": ...}

```

## bench results sequential

Kept the worker sleep time as 0.1 sec when no job available

```bash
python bench.py
Submitting 100 jobs....
Benchmark complete.
Average latency: 0.10369821786880493
P50: 0.1032710075378418
P90: 0.10610103607177734
P99: 0.10786104202270508
```

## bench 2 parallel

Kept worker sleep time as 2 sec. Because parallely jobs are getting submitted, worker doesn't sleep that much

```
python bench2.py
Submitted 100 jobs with 20 concurrent clients.
Total time: 7.004s
Throughput: 14.28 jobs/sec
```

“Long polling” is meant for clients waiting for events, not workers pulling jobs
Long-polling evolved as a browser hack before WebSockets existed.
Typical users:
browsers waiting for chat messages
mobile clients waiting for notifications
dashboards waiting for job updates
Workers are not “clients” in that sense.
Workers are producers/consumers in a job system.
Workers should not behave like UI clients.

Queues are the correct worker mechanism, long polling, web sockets, and sse are for idle clients.

## long polling

- its a hybrid. Worker sends GET, Server holds the connection open upto 30,60 sec.
- when a job arrives server responds immediately
- if time expires server returns 204

- so the worker makes 1 request every 60 seconds

Used by Kubernetes watch API, etc.

## WebSockets

- persistent duplex connection
- server -> worker: and worker->server
- no open port needed on worker

Just some notes from gpt

We’ll cover:

1. **Does long polling save CPU?**
2. **How does a server handle multiple long-polling TCP connections?**
3. **Why long-polling workers cause backpressure & deadlocks** (the real danger)
4. **Why long polling is great for _clients_ but harmful for _workers_**.

Let’s go step by step.

---

# 1. **Does long polling save CPU?**

### ✔ Yes — long polling saves CPU compared to short polling.

Let’s compare:

### ❌ Short Polling worker (bad)

```
loop:
    GET /next_job       → returns 204
    sleep(100ms)
```

This wastes:

- 10 requests/second per worker
- server continuously decoding HTTP
- worker doing repeated request/response

### ✔ Long Polling worker (good _in this narrow sense_)

```
GET /next_job → server waits 20 seconds
```

While waiting:

- no CPU work
- goroutine blocks on `select`
- worker thread sleeps internally in requests library
- network socket idle

### So yes:

**long polling reduces CPU load** vs short polling.

---

# 2. **How does the server handle multiple connections?**

You asked:

> "can't it connect to multiple?"

### ✔ Yes — a Go server can handle **many** long polling connections.

In Go:

```
each incoming HTTP request = one goroutine
```

And goroutines are extremely lightweight:

- ~2 KB stack
- multiplexed on OS threads
- scheduler parks blocked goroutines with near-zero cost

So this is fine:

```
1000 clients doing long polling
=> 1000 goroutines
=> almost no CPU usage
```

This is exactly why Go is fantastic for long polling.

### So long polling itself is NOT the bottleneck.

**The issue is not Go — the issue is using long-poll incorrectly in a job system.**

---

# 3. **What does the server actually do during long polling?**

During long polling:

```
client --TCP--> server
server goroutine blocks on select:
    case event:
    case <-timeout:
```

So the socket is open but idle.
The goroutine is sleeping.
CPU usage is near zero.

**TCP connection = not expensive**
**Goroutine = almost free**

So long-polling is great — _when the client is a UI or human-facing consumer._

---

# 4. **Then why not let Python workers long poll?**

Perfect question.
This is where the _architectural problem_ begins.

Let’s explain the backpressure deadlock.

---

# 5. 🔥 **Backpressure Deadlock Explanation (The Dangerous Part)**

Let’s model your architecture:

```
Workers → GET /next_job  (long polling)
Server → returns job when available
```

Now imagine:

### **Case: 5 workers, 1 job arrives**

- All 5 workers call `/next_job`
- All 5 connections are long-polling
- Server picks 1 job and returns it to _one_ worker
- 4 workers remain hanging in long-poll
- Server is holding 4 idle TCP connections

Now imagine:

### **More workers appear: 500 workers**

- They all long-poll `/next_job`
- Server holds 500 open connections
- No jobs at the moment
- Workers waiting indefinitely

Still fine.

Now here's the twist:

---

# 🧠 **The Deadlock Scenario**

Imagine the server ONLY sends a job when a worker "asks".

This requires:

```
Worker asks → server pops job → server responds
```

BUT long-poll workers behave like this:

```
Worker asks once → waits 20 seconds → retry
```

### ⚠️ The moment all workers are waiting (inside long poll), **no new worker requests are made**.

So:

- Server gets a job → but no new `/next_job` requests come in
- Because all existing workers are busy waiting on older requests
- Server has NO active workers asking for jobs
- Server can't dispatch the new job
- Workers continue holding sockets but NOT asking again

➡ **This creates backpressure** (server has jobs but nobody is "pulling").
➡ **Eventually becomes a deadlock**.

### **WHY?**

Because long polling turns workers into **passive listeners**, not **active job requesters**.

Workers block on _old_ requests, but the job that arrives requires a _new_ request.

So new jobs can ONLY be assigned AFTER worker long-poll times out.

This adds huge delays.

Now scale up:

- 10 workers → mild
- 100 workers → seconds delay
- 1000 workers → minutes of delay
- 10,000 workers → near-permanent stuck state

### ❗ This is why long polling breaks job systems.

It **disconnects job dispatching from worker demand**.

---

# 6. 🔥 Even Worse: Long Poll Inverts Control (Wrong Direction)

Correct producer–consumer pattern:

```
Worker: I am ready, give me a job.
Server: OK, here you go.
```

Long polling flips it into:

```
Worker: I am waiting for something…
Server: I will notify you when something happens.
```

That is an **event subscription model**, not a **job queue model**.

Your system is supposed to be:

- Worker pulls jobs
- Server does not push jobs

Long polling tries to turn it into:

- Server pushes jobs
- Worker waits passively

This breaks everything.

---

# 7. 🧠 “Why doesn't long polling break UI apps then?”

UI clients (browser, mobile, CLI) do NOT:

- pop jobs from queues
- require fairness
- need to keep throughput
- scale to thousands
- need to maintain job dispatch order

They wait for _notifications_, not _work_.

Workers cannot behave like that.

---

# 8. ✔ Correct patterns for Workers

### Best

```
Redis BRPOP
```

### Good

```
RabbitMQ / NATS / Kafka consumer
```

### Acceptable

```
Short polling every 2 seconds
```

### Dangerous

```
Long polling
```

---

# 🎯 Final Summary

### ✔ Long polling _is_ CPU efficient

### ✔ Go can handle thousands of long-poll TCP connections

### ✔ But long-polling WORKERS is architecturally **wrong** and causes:

- **Backpressure:** server has jobs but no workers actively asking
- **Deadlock:** all workers stuck waiting for earlier polls
- **Stuck jobs:** jobs only delivered after long-poll timeout
- **Massive latency:** job isn’t picked immediately
- **Wrong control flow:** pushes notification model into a pull-based system
- **Scalability disaster:** 1000 workers → 1000 idle sockets

### ✔ Long polling works perfectly for CLIENTS

(e.g., `/status?id=`)

### ❌ Long polling workers = bad practice

(breaks pull-based job systems)
