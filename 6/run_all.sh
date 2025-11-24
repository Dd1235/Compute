#!/bin/bash
# run_all.sh — start and stop Go server, Python worker, and Node clients cleanly

set -e

echo "Cleaning up old processes..."
pkill -f "main.go" 2>/dev/null || true
pkill -f "worker.py" 2>/dev/null || true
pkill -f "client.js" 2>/dev/null || true
redis-cli FLUSHDB >/dev/null 2>&1 || true

echo "Environment cleaned."

echo "Starting Go server..."
(cd server && exec go run main.go) &
SERVER_PID=$!
sleep 3  

echo "Starting Python worker..."
(cd worker && exec python worker.py) &
WORKER_PID=$!
sleep 1   # wait for worker to connect to Redis

echo "Starting Node clients..."
(cd client && exec node client.js 1) &
(cd client && exec node client.js 2) &
(cd client && exec node client.js 3) &


echo "All components running."
echo "Press Ctrl+C to stop everything."

# trap Ctrl+C and terminate all stored PIDs
trap 'echo "Stopping everything..."; \
      kill $SERVER_PID $WORKER_PID $CLIENT1_PID $CLIENT2_PID $CLIENT3_PID 2>/dev/null; exit 0' SIGINT

wait
