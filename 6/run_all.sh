#!/bin/bash
# ==========================================================
# run_all.sh
# Starts Go server, Python worker, and Node client together
# ==========================================================

set -e  # exit if any command fails

SERVER_FILE="./server/main.go"
WORKER_FILE="./worker/worker.py"
CLIENT_FILE="./client/client.js"

echo "Checking dependencies..."
command -v go >/dev/null 2>&1 || { echo "Go not installed"; exit 1; }
command -v python >/dev/null 2>&1 || { echo "Python not installed"; exit 1; }
command -v node >/dev/null 2>&1 || { echo "Node.js not installed"; exit 1; }

echo "Assuming Redis is already running at localhost:6379"

echo "Starting Go server..."
(cd ./server && go run main.go) &
SERVER_PID=$!
sleep 1

echo "Starting Python worker..."
(cd ./worker && python worker.py) &
WORKER_PID=$!
sleep 1

echo "Starting Node client..."
(cd ./client && node client.js) &
CLIENT_PID=$!

echo "All components running."
echo "Press Ctrl+C to stop everything."

trap 'echo "Stopping..."; kill $SERVER_PID $WORKER_PID $CLIENT_PID 2>/dev/null; exit 0' INT

wait
