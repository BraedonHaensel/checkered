#!/usr/bin/env bash

MATCHMAKERS=${1:-3}
GAME_SERVERS=${2:-3}

BASE_PORT_MATCHMAKER=4000
BASE_PORT_GAMESERVER=5000
NS_ADDR="http://localhost:9000"

PIDS=()

prefix_output() {
  local prefix=$1
  while IFS= read -r line; do
    echo "[$prefix] $line"
  done
}

cleanup() {
  echo "Shutting down..."
  for pid in "${PIDS[@]}"; do
    kill "$pid" 2>/dev/null
  done
  wait
  exit 0
}

trap cleanup SIGINT SIGTERM

# Frontend
(
    cd frontend || exit 1
    npm run dev 2>&1 | prefix_output "frontend"
) &
PIDS+=($!)

# Name server
(
  cd name-server || exit 1
  go run . 2>&1 | prefix_output "name-server:1"
) &
PIDS+=($!)

sleep 2

# Matchmakers
for ((i=1; i<=MATCHMAKERS; i++)); do
  PORT=$((BASE_PORT_MATCHMAKER + i))

  (
    cd backend || exit 1
    go run cmd/matchmaker/main.go \
      --addr ":$PORT" \
      -ns "$NS_ADDR" 2>&1 | prefix_output "matchmaker:$i"
  ) &

  PIDS+=($!)
done

# Game servers
for ((i=1; i<=GAME_SERVERS; i++)); do
  PORT=$((BASE_PORT_GAMESERVER + i))

  (
    cd backend || exit 1
    go run cmd/game-server/main.go \
      --addr ":$PORT" \
      -ns "$NS_ADDR" 2>&1 | prefix_output "game-server:$i"
  ) &

  PIDS+=($!)
done

wait
