#!/usr/bin/env bash
#
# End-to-end test: starts 3 backends + the LB, then exercises every balancing
# algorithm by patching config.yaml, restarting the LB, and counting how many
# requests each backend served. Cleans up (kills processes, restores config) on exit.
#
# Usage:  ./scripts/test.sh            # test all algorithms
#         ./scripts/test.sh weighted   # test just one
#
set -eu

LB="localhost:8080"
PORTS=(9001 9002 9003)
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

ALGOS=(round_robin weighted least_conn ip_hash consistent_hash)
[ $# -ge 1 ] && ALGOS=("$1")

echo "building binaries..."
go build -o ./bin/backend ./cmd/backend
go build -o ./bin/lb ./cmd/lb

# back up config so we can restore it after mutating the algorithm line
CFG_BACKUP="$(mktemp)"
cp config.yaml "$CFG_BACKUP"

BACKEND_PIDS=()
LB_PID=""

cleanup() {
  echo
  echo "cleaning up..."
  [ -n "$LB_PID" ] && kill "$LB_PID" 2>/dev/null || true
  for pid in "${BACKEND_PIDS[@]}"; do kill "$pid" 2>/dev/null || true; done
  cp "$CFG_BACKUP" config.yaml && rm -f "$CFG_BACKUP"
}
trap cleanup EXIT

wait_lb() {
  for _ in $(seq 1 30); do
    curl -s -o /dev/null "$LB/" 2>/dev/null && return 0
    sleep 0.2
  done
  return 1
}

start_lb() {
  ./bin/lb >/dev/null 2>&1 &
  LB_PID=$!
  wait_lb || { echo "!! LB failed to start (bad config?)"; exit 1; }
}

stop_lb() {
  [ -n "$LB_PID" ] && kill "$LB_PID" 2>/dev/null || true
  wait "$LB_PID" 2>/dev/null || true
  LB_PID=""
}

tally() { grep -oE 'port [0-9]+' | sort | uniq -c; }

# start the backends once; they stay up across all algorithm runs
for p in "${PORTS[@]}"; do
  ./bin/backend -port="$p" >/dev/null 2>&1 &
  BACKEND_PIDS+=($!)
done
sleep 0.5

for algo in "${ALGOS[@]}"; do
  sed -i -E "s/^algorithm:.*/algorithm: ${algo}/" config.yaml
  start_lb
  echo
  echo "===== algorithm: ${algo} ====="

  case "$algo" in
    round_robin|weighted)
      # sequential requests — distribution shows the ordering/weighting
      for _ in $(seq 1 30); do curl -s "$LB/"; done | tally
      ;;
    least_conn)
      # needs concurrency + latency to differ from RR — hit /slow in parallel
      tmp="$(mktemp)"
      for _ in $(seq 1 12); do curl -s "$LB/slow" >>"$tmp" & done
      wait
      tally < "$tmp"
      rm -f "$tmp"
      ;;
    ip_hash|consistent_hash)
      # distinct clients via spoofed X-Forwarded-For
      for ip in 1.1.1.1 2.2.2.2 3.3.3.3 4.4.4.4 5.5.5.5 6.6.6.6 7.7.7.7 8.8.8.8; do
        curl -s -H "X-Forwarded-For: $ip" "$LB/"
      done | tally
      ;;
  esac

  stop_lb
done

echo
echo "done — all requested algorithms tested."
