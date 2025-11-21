import hashlib
import time

import requests

SERVER = "http://localhost:8080"


def sha256_hex(s: str) -> str:
    return hashlib.sha256(s.encode()).hexdigest()


while True:
    r = requests.get(f"{SERVER}/next_job")
    if r.status_code == 204:
        print("No jobs. Sleeping...")
        time.sleep(2)
        continue

    job = r.json()
    print("Processing job:", job["id"], job["input"])

    result = sha256_hex(job["input"])

    requests.post(f"{SERVER}/job_result", json={"id": job["id"], "result": result})

    print("Completed:", job["id"])
