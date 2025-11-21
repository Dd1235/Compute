import threading
import time
import uuid

import requests

SERVER = "http://localhost:8080"

# we are only measuring sequential throughput not concurrent load with this one
# end to end latency without queue pressure


def submit_job(input_data: str) -> str:
    r = requests.post(f"{SERVER}/submit", json={"input": input_data})
    return r.json()["id"]


def wait_for_result(job_id):
    while True:
        r = requests.get(f"{SERVER}/status?id={job_id}")
        if r.status_code == 200:
            j = r.json()
            if j["status"] == "done":
                return j
    time.sleep(0.05)


def run_benchmark(num_jobs=100):
    job_ids = []
    latencies = []
    print(f"Submitting {num_jobs} jobs....")

    for _ in range(num_jobs):
        t0 = time.time()
        job_id = submit_job("hello world")
        job_ids.append(job_id)
        result = wait_for_result(job_id)
        t1 = time.time()
        latencies.append(t1 - t0)

    print("Benchmark complete.")

    print("Average latency: ", sum(latencies) / len(latencies))
    print("P50:", sorted(latencies)[len(latencies) // 2])
    print("P90:", sorted(latencies)[int(len(latencies) * 0.9)])
    print("P99:", sorted(latencies)[int(len(latencies) * 0.99)])


if __name__ == "__main__":
    run_benchmark(100)
