import json
import time

import redis

# connect to the same redis instance
r = redis.Redis(host="localhost", port=6379, db=0)

print("Worker started... waiting for jobs.")


def process_job(job):
    job_id = job["id"]
    for pct in range(0, 101, 20):
        time.sleep(1)
        update = {"job_id": job_id, "status": f"progress: {pct}%", "result": ""}
        r.hset(f"job:{job_id}", mapping={"status": update["status"]})
        r.publish("job_updates", json.dumps(update))

    result = job["input"].upper()
    update = {"job_id": job_id, "status": "done", "result": result}
    r.hset(f"job:{job['id']}", mapping={"status": "done", "result": result})
    r.publish("job_updates", json.dumps(update))


while True:

    # atomic retrival and deletion
    # there is no build in linkage between the q and hashes
    _, job_data = r.brpop("jobs")  # blocks until new job
    job = json.loads(job_data)
    print(f"Processing job {job['id']} -> {job['input']}")
    process_job(job)
