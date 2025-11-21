import threading
import time
import uuid
from concurrent.futures import ThreadPoolExecutor

import requests

SERVER = "http://localhost:8080"


def submit_job(input_data):
    r = requests.post(f"{SERVER}/submit", json={"input": input_data})
    return r.json()["id"]


def wait_for_result(job_id):
    while True:
        r = requests.get(f"{SERVER}/status?id={job_id}")
        if r.status_code == 200 and r.json()["status"] == "done":
            return
        time.sleep(0.05)


def submit_and_wait(i):
    job_id = submit_job(f"payload-{i}")
    wait_for_result(job_id)


# === Parallel benchmark ===
def run_parallel(n=100, workers=20):
    t0 = time.time()

    with ThreadPoolExecutor(max_workers=workers) as pool:
        pool.map(submit_and_wait, range(n))

    t1 = time.time()

    print(f"Submitted {n} jobs with {workers} concurrent clients.")
    print(f"Total time: {t1 - t0:.3f}s")
    print(f"Throughput: {n / (t1 - t0):.2f} jobs/sec")


if __name__ == "__main__":
    run_parallel(100)
