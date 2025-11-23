# Exercise 6

`LPUSH queue job1` -> pushes a job to the left of the q (q head)
`BRPOP queue 0` -> blocks until a job appears, then pops the last element(q tail)

- server should add jobs to the queue, and worker should retrieve jobs

- Let's refactor 5/ to use Redis instead of in memory list in the server.
  This way: - job doesn't vanish if server restarts - mulitple workers can share the same q safely - workers don't have to poll the /next_job endpoint on hte server - there is no manual locking - it is persistent - workers wakeup instatnly as new jobs appear

Set up redis either using docker `docker run -d -p 6379:6379 --name redis redis:7` or natively.

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
to properly stop as it might be running as homebrew server
brew services stop redis
brew services start redis
```
