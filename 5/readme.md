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

```

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
```
