import json
import threading
import time
import uuid

import psutil
import redis
import requests

WORKER_ID = f"worker-{uuid.uuid4().hex[:6]}"
REDIS = redis.Redis(host="localhost", port=6379, db=0)

# generate a random worker id, connect to the regis instance, make a function that registers you with the go server with the cpu ad memory, and makes the post request, heart beat function sets the last_heartbeat associated with that worker id in redis every 5 seconds, in the mail loop, if there is a job assigned to this worker, dequeu that, get the json for the job, simulate some work, set the job as done and set its result, and set your status as idle

# brpop - blocking right pop, so each worker only listens to its personal queue


def register():
    data = {
        "id": WORKER_ID,
        "cpu": psutil.cpu_count(),
        "memory_gb": int(psutil.virtual_memory().total / 1e9),
    }
    print("Registering worker", WORKER_ID)
    requests.post("http://localhost:8080/register", json=data)


def heartbeat():
    while True:
        REDIS.hset(f"worker:{WORKER_ID}", "last_heartbeat", int(time.time()))
        time.sleep(5)


def main_loop():
    while True:
        channel = f"worker:{WORKER_ID}:jobs"
        _, job_json = REDIS.brpop(channel)
        job = json.loads(job_json)
        print(f"Worker {WORKER_ID} picked job {job['id']}")
        REDIS.hset(f"job:{job['id']}", "status", "running")
        time.sleep(2)  # simulate work
        result = job["input"].upper()
        REDIS.hset(f"job:{job['id']}", mapping={"status": "done", "result": result})
        REDIS.hset(f"worker:{WORKER_ID}", "status", "idle")
        print(f"Job {job['id']} done: {result}")


if __name__ == "__main__":
    register()
    threading.Thread(target=heartbeat, daemon=True).start()
    main_loop()
