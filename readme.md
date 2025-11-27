Some simple exercise (as simple as making a http server) and their solutions.

todo:

- [] Need to explore how event loops and go routines work.
- [] more on redis queue
- [] writing benchmarks and metrics

Folders:

- 1/: Minimal Go HTTP server with hello, headers echo, and ping endpoints.
- 2/: Python client examples hitting the Go server via requests and http.client.
- 3/: Go server that hashes submitted names via JSON POST and returns the SHA-256.
- 4/: Go playground using go-redis to store/fetch a sample struct in Redis.
- 5/: In-memory job queue server with short-polling worker/client benchmarks.
- 6/: Redis-backed job queue with worker, client, and server using BRPOP and websockets for updates.
- 7/: Redis-based scheduler that registers workers, assigns jobs to per-worker queues, and tracks status.
- 8/: Simple Go map-reduce word counter using goroutines for mappers/reducers over an input file.
