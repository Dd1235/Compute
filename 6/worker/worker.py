import json
import time

import redis

# connect to the same redis instance
r = redis.Redis(host="localhost", port=6379, db=0)

print("Worker started... waiting for jobs.")

while True:

    # atomic retrival and deletion
    # there is no build in linkage between the q and hashes
    _, job_data = r.brpop("jobs")  # blocks until new job
    job = json.loads(job_data)
    print(f"Processing job {job['id']} -> {job['input']}")

    result = job["input"].upper()
    time.sleep(2)

    r.hset(f"job:{job['id']}", mapping={"status": "done", "result": result})
    print(f"Done! job {job['id']} = {result}")
